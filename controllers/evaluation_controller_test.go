package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"cybersim/dto"
	"cybersim/middlewares"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type fakeEvalStore struct {
	mu                 sync.Mutex
	exams              map[string][]*models.ExamSession
	metricsByUser      map[string]*models.EvaluationMetric
	profilesByUser     map[string][]models.UserSkillProfile
	completedScenarios map[string]int
	userRoles          map[string]string
	userNicknames      map[string]string
	nextEvalSeq        int
}

func newFakeEvalStore() *fakeEvalStore {
	return &fakeEvalStore{
		exams:              make(map[string][]*models.ExamSession),
		metricsByUser:      make(map[string]*models.EvaluationMetric),
		profilesByUser:     make(map[string][]models.UserSkillProfile),
		completedScenarios: make(map[string]int),
		userRoles:          make(map[string]string),
		userNicknames:      make(map[string]string),
	}
}

func (f *fakeEvalStore) FindLatestCompletedExam(userID, examType string) (*models.ExamSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var latest *models.ExamSession
	for _, ex := range f.exams[userID] {
		if ex.ExamType != examType || ex.Status != "completed" {
			continue
		}
		if latest == nil {
			latest = ex
			continue
		}
		if ex.CompletedAt != nil && (latest.CompletedAt == nil || ex.CompletedAt.After(*latest.CompletedAt)) {
			latest = ex
		}
	}
	if latest == nil {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *latest
	return &cp, nil
}

func (f *fakeEvalStore) FindEvaluationMetricByUser(userID string) (*models.EvaluationMetric, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.metricsByUser[userID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *m
	return &cp, nil
}

func (f *fakeEvalStore) GetLearnerTrainingSnapshot(userID string) (int, []models.UserSkillProfile, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	profiles := append([]models.UserSkillProfile(nil), f.profilesByUser[userID]...)
	return f.completedScenarios[userID], profiles, nil
}

func (f *fakeEvalStore) CreateEvaluationMetricAtomic(metric *models.EvaluationMetric) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.metricsByUser[metric.UserID]; exists {
		return ErrQuestionnaireAlreadySubmitted
	}
	f.nextEvalSeq++
	if metric.EvalID == "" {
		metric.EvalID = "eval-uuid-" + string(rune('0'+f.nextEvalSeq))
	}
	if metric.EvalDate.IsZero() {
		metric.EvalDate = time.Now().UTC()
	}
	cp := *metric
	f.metricsByUser[metric.UserID] = &cp
	return nil
}

func (f *fakeEvalStore) GetAdminLearnerEvaluations() ([]services.AdminLearnerEvaluationRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]services.AdminLearnerEvaluationRecord, 0, len(f.metricsByUser))
	for uid, m := range f.metricsByUser {
		role := f.userRoles[uid]
		if role != "" && role != "learner" {
			continue
		}
		nick := f.userNicknames[uid]
		if nick == "" {
			nick = uid
		}
		out = append(out, services.AdminLearnerEvaluationRecord{
			UserID:   uid,
			Nickname: nick,
			Metric:   *m,
		})
	}
	return out, nil
}

func (f *fakeEvalStore) GetLearnerCohortFunnel() (dto.EvaluationCohortFunnelDTO, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	totalLearners := 0
	preDone := 0
	postDone := 0
	for uid, role := range f.userRoles {
		if role != "learner" {
			continue
		}
		totalLearners++
		hasPre := false
		hasPost := false
		for _, ex := range f.exams[uid] {
			if ex.Status == "completed" && ex.ExamType == "pre" {
				hasPre = true
			}
			if ex.Status == "completed" && ex.ExamType == "post" {
				hasPost = true
			}
		}
		if hasPre {
			preDone++
		}
		if hasPost {
			postDone++
		}
	}
	respondents := 0
	for uid := range f.metricsByUser {
		role := f.userRoles[uid]
		if role == "" || role == "learner" {
			respondents++
		}
	}
	rate := 0.0
	if postDone > 0 {
		rate = roundMetric2((float64(respondents) / float64(postDone)) * 100.0)
	}
	return dto.EvaluationCohortFunnelDTO{
		TotalLearners:             totalLearners,
		CompletedPreTestLearners:  preDone,
		CompletedPostTestLearners: postDone,
		QuestionnaireRespondents:  respondents,
		ResponseRatePct:           rate,
	}, nil
}

func setupEvalTestRouter(ec *EvaluationController, userID, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authInject := func(c *gin.Context) {
		if userID != "" {
			c.Set("user_id", userID)
		}
		if role != "" {
			c.Set("role", role)
		}
		c.Next()
	}
	r.GET("/api/evaluation/questionnaire/status", authInject, ec.GetQuestionnaireStatus)
	r.POST("/api/evaluation/questionnaire", authInject, ec.SubmitQuestionnaire)
	admin := r.Group("/api/admin", authInject, middlewares.RequireAdmin())
	{
		admin.GET("/evaluations", ec.GetAdminEvaluations)
		admin.GET("/evaluations/statistics", ec.GetAdminStatistics)
		admin.GET("/evaluations/export", ec.ExportAdminEvaluations)
	}
	return r
}

func makeValidQuestionnaireReqBody() string {
	return `{
		"locale": "th",
		"csuq": {
			"q1": 6, "q2": 7, "q3": 6, "q4": 6, "q5": 7, "q6": 6,
			"q7": null, "q8": 6, "q9": 6, "q10": 6, "q11": 7, "q12": 6,
			"q13": 7, "q14": 6, "q15": 6, "q16": 7
		},
		"nasa_tlx": {
			"mental_demand": 60,
			"physical_demand": 10,
			"temporal_demand": 45,
			"performance_rating": 85,
			"effort": 55,
			"frustration": 20
		},
		"qualitative_feedback": {
			"most_impressive": "SIEM logs felt realistic",
			"areas_to_improve": "More cloud scenarios",
			"additional_comments": "Great platform"
		}
	}`
}

func seedCompletedPreAndPost(store *fakeEvalStore, userID string, preScore, postScore float64) {
	now := time.Now().UTC()
	preDone := now.Add(-2 * time.Hour)
	postDone := now.Add(-5 * time.Minute)
	store.exams[userID] = []*models.ExamSession{
		{SessionID: "pre-1", UserID: userID, ExamType: "pre", Status: "completed", Score: &preScore, CompletedAt: &preDone},
		{SessionID: "post-1", UserID: userID, ExamType: "post", Status: "completed", Score: &postScore, CompletedAt: &postDone},
	}
	store.completedScenarios[userID] = 4
	store.profilesByUser[userID] = []models.UserSkillProfile{
		{UserID: userID, DomainID: "domain1", ProficiencyScore: 80.0},
		{UserID: userID, DomainID: "domain2", ProficiencyScore: 74.5},
		{UserID: userID, DomainID: "domain3", ProficiencyScore: 82.0},
		{UserID: userID, DomainID: "domain4", ProficiencyScore: 75.5},
		{UserID: userID, DomainID: "domain5", ProficiencyScore: 80.0},
	}
}

func TestQuestionnaireStatus_RolesAndEligibility(t *testing.T) {
	store := newFakeEvalStore()
	ec := NewEvaluationControllerWithStore(store)

	// 1. Learner before Post-Test -> not eligible
	rPre := setupEvalTestRouter(ec, "learner-1", "learner")
	req := httptest.NewRequest(http.MethodGet, "/api/evaluation/questionnaire/status", nil)
	w := httptest.NewRecorder()
	rPre.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var st dto.QuestionnaireStatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &st); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if st.Eligible || st.IneligibleReason != "post_test_required" || st.IsPreviewOnly || st.Submitted {
		t.Fatalf("unexpected status before post-test: %+v", st)
	}

	// 2. Specialist/Admin -> Preview Mode only
	rSpec := setupEvalTestRouter(ec, "spec-1", "specialist")
	wSpec := httptest.NewRecorder()
	rSpec.ServeHTTP(wSpec, httptest.NewRequest(http.MethodGet, "/api/evaluation/questionnaire/status", nil))
	var stSpec dto.QuestionnaireStatusResponse
	_ = json.Unmarshal(wSpec.Body.Bytes(), &stSpec)
	if !stSpec.IsPreviewOnly || stSpec.Eligible || stSpec.IneligibleReason != "preview_only_non_learner" {
		t.Fatalf("unexpected specialist preview status: %+v", stSpec)
	}

	// 3. Learner after Post-Test -> eligible
	seedCompletedPreAndPost(store, "learner-1", 60.0, 83.33)
	wAfter := httptest.NewRecorder()
	rPre.ServeHTTP(wAfter, httptest.NewRequest(http.MethodGet, "/api/evaluation/questionnaire/status", nil))
	var stAfter dto.QuestionnaireStatusResponse
	_ = json.Unmarshal(wAfter.Body.Bytes(), &stAfter)
	if !stAfter.Eligible || !stAfter.HasCompletedPostTest || stAfter.Submitted {
		t.Fatalf("expected eligible learner after post-test: %+v", stAfter)
	}
}

func TestSubmitQuestionnaire_GuardsSnapshotAndDuplicate(t *testing.T) {
	store := newFakeEvalStore()
	ec := NewEvaluationControllerWithStore(store)

	// 1. Non-learner rejected with 403 QUESTIONNAIRE_LEARNER_ONLY
	rAdmin := setupEvalTestRouter(ec, "admin-1", "admin")
	wAdmin := httptest.NewRecorder()
	rAdmin.ServeHTTP(wAdmin, httptest.NewRequest(http.MethodPost, "/api/evaluation/questionnaire", bytes.NewBufferString(makeValidQuestionnaireReqBody())))
	if wAdmin.Code != http.StatusForbidden || !bytes.Contains(wAdmin.Body.Bytes(), []byte("QUESTIONNAIRE_LEARNER_ONLY")) {
		t.Fatalf("expected 403 QUESTIONNAIRE_LEARNER_ONLY, got %d: %s", wAdmin.Code, wAdmin.Body.String())
	}

	// 2. Learner without Post-Test rejected with 403 QUESTIONNAIRE_POST_TEST_REQUIRED
	rLearner := setupEvalTestRouter(ec, "learner-1", "learner")
	wNoPost := httptest.NewRecorder()
	rLearner.ServeHTTP(wNoPost, httptest.NewRequest(http.MethodPost, "/api/evaluation/questionnaire", bytes.NewBufferString(makeValidQuestionnaireReqBody())))
	if wNoPost.Code != http.StatusForbidden || !bytes.Contains(wNoPost.Body.Bytes(), []byte("QUESTIONNAIRE_POST_TEST_REQUIRED")) {
		t.Fatalf("expected 403 QUESTIONNAIRE_POST_TEST_REQUIRED, got %d: %s", wNoPost.Code, wNoPost.Body.String())
	}

	// 3. Seed Pre/Post + Profiles (proficiency scores must remain untouched)
	seedCompletedPreAndPost(store, "learner-1", 60.0, 83.33)
	wCreate := httptest.NewRecorder()
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/evaluation/questionnaire", bytes.NewBufferString(makeValidQuestionnaireReqBody()))
	reqCreate.Header.Set("Content-Type", "application/json")
	rLearner.ServeHTTP(wCreate, reqCreate)
	if wCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", wCreate.Code, wCreate.Body.String())
	}

	var resp dto.SubmitQuestionnaireResponse
	if err := json.Unmarshal(wCreate.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal submit response: %v", err)
	}
	sub := resp.Submission
	if sub.PreScore != 60.0 || sub.PostScore != 83.33 || sub.ImprovementPct != 23.33 {
		t.Fatalf("unexpected server snapshot scores: pre=%v post=%v imp=%v", sub.PreScore, sub.PostScore, sub.ImprovementPct)
	}
	if sub.CompletedScenarios != 4 || sub.AverageProficiency != 78.4 {
		t.Fatalf("unexpected training snapshot: scenarios=%d avgProf=%v", sub.CompletedScenarios, sub.AverageProficiency)
	}
	if sub.CSUQScores.Subscales.Overall == nil || *sub.CSUQScores.Subscales.Overall != 6.33 {
		t.Fatalf("unexpected CSUQ overall: %v", sub.CSUQScores.Subscales.Overall)
	}
	if sub.NASATLXScores.Dimensions.PerformanceWorkload != 15 || sub.NASATLXScores.OverallWorkload != 34.17 {
		t.Fatalf("unexpected TLX scores: %+v", sub.NASATLXScores)
	}
	// Proficiency invariance check
	if store.profilesByUser["learner-1"][0].ProficiencyScore != 80.0 {
		t.Fatalf("proficiency score mutated unexpectedly")
	}

	// 4. Duplicate submission rejected with 409 QUESTIONNAIRE_ALREADY_SUBMITTED
	wDup := httptest.NewRecorder()
	reqDup := httptest.NewRequest(http.MethodPost, "/api/evaluation/questionnaire", bytes.NewBufferString(makeValidQuestionnaireReqBody()))
	reqDup.Header.Set("Content-Type", "application/json")
	rLearner.ServeHTTP(wDup, reqDup)
	if wDup.Code != http.StatusConflict || !bytes.Contains(wDup.Body.Bytes(), []byte("QUESTIONNAIRE_ALREADY_SUBMITTED")) {
		t.Fatalf("expected 409 QUESTIONNAIRE_ALREADY_SUBMITTED, got %d: %s", wDup.Code, wDup.Body.String())
	}

	// 5. Status now reflects submitted=true with read-only submission payload
	wStatus := httptest.NewRecorder()
	rLearner.ServeHTTP(wStatus, httptest.NewRequest(http.MethodGet, "/api/evaluation/questionnaire/status", nil))
	var stDone dto.QuestionnaireStatusResponse
	_ = json.Unmarshal(wStatus.Body.Bytes(), &stDone)
	if !stDone.Submitted || stDone.Eligible || stDone.IneligibleReason != "already_submitted" || stDone.Submission == nil {
		t.Fatalf("unexpected submitted status: %+v", stDone)
	}

	// 6. Missing/nil NASA-TLX dimension rejects with 400 QUESTIONNAIRE_INVALID_TLX (not INVALID_CSUQ)
	seedCompletedPreAndPost(store, "learner-tlx-miss", 50.0, 80.0)
	rTlxMiss := setupEvalTestRouter(ec, "learner-tlx-miss", "learner")
	missingTlxPayload := `{
		"locale": "en",
		"csuq": {"q1":6,"q2":7,"q3":6,"q4":6,"q5":7,"q6":6,"q7":6,"q8":6,"q9":6,"q10":6,"q11":7,"q12":6,"q13":7,"q14":6,"q15":6,"q16":7},
		"nasa_tlx": {"mental_demand":60,"physical_demand":10,"temporal_demand":45,"performance_rating":85,"effort":55}
	}`
	wTlx := httptest.NewRecorder()
	reqTlx := httptest.NewRequest(http.MethodPost, "/api/evaluation/questionnaire", bytes.NewBufferString(missingTlxPayload))
	reqTlx.Header.Set("Content-Type", "application/json")
	rTlxMiss.ServeHTTP(wTlx, reqTlx)
	if wTlx.Code != http.StatusBadRequest || !bytes.Contains(wTlx.Body.Bytes(), []byte("QUESTIONNAIRE_INVALID_TLX")) {
		t.Fatalf("expected 400 QUESTIONNAIRE_INVALID_TLX for omitted dimension, got %d: %s", wTlx.Code, wTlx.Body.String())
	}
}

func TestAdminEvaluations_RBACPseudonymizationAndExport(t *testing.T) {
	store := newFakeEvalStore()
	ec := NewEvaluationControllerWithStore(store)

	uid := "abcdef12-3456-7890-abcd-ef1234567890"
	store.userRoles[uid] = "learner"
	store.userNicknames[uid] = "blue_hero"
	seedCompletedPreAndPost(store, uid, 55.0, 85.0)

	// Submit as learner with formula-like text in notes to verify CSV safety
	rLearner := setupEvalTestRouter(ec, uid, "learner")
	payload := `{
		"locale": "th",
		"csuq": {"q1":6,"q2":7,"q3":6,"q4":6,"q5":7,"q6":6,"q7":null,"q8":6,"q9":6,"q10":6,"q11":7,"q12":6,"q13":7,"q14":6,"q15":6,"q16":7},
		"nasa_tlx": {"mental_demand":60,"physical_demand":10,"temporal_demand":45,"performance_rating":85,"effort":55,"frustration":20},
		"qualitative_feedback": {"most_impressive":"=SUM(A1:A2)", "areas_to_improve":"", "additional_comments":"ดีมาก"}
	}`
	wSub := httptest.NewRecorder()
	reqSub := httptest.NewRequest(http.MethodPost, "/api/evaluation/questionnaire", bytes.NewBufferString(payload))
	reqSub.Header.Set("Content-Type", "application/json")
	rLearner.ServeHTTP(wSub, reqSub)
	if wSub.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", wSub.Code, wSub.Body.String())
	}

	// Non-admin RBAC checks (learner and specialist must get 403)
	for _, role := range []string{"learner", "specialist"} {
		rForbidden := setupEvalTestRouter(ec, "other", role)
		wF := httptest.NewRecorder()
		rForbidden.ServeHTTP(wF, httptest.NewRequest(http.MethodGet, "/api/admin/evaluations", nil))
		if wF.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for role %s on /evaluations, got %d", role, wF.Code)
		}
		wExportF := httptest.NewRecorder()
		rForbidden.ServeHTTP(wExportF, httptest.NewRequest(http.MethodGet, "/api/admin/evaluations/export?format=csv", nil))
		if wExportF.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for role %s on /export, got %d", role, wExportF.Code)
		}
	}

	// Admin evaluations fetch
	rAdmin := setupEvalTestRouter(ec, "admin-1", "admin")
	wAdmin := httptest.NewRecorder()
	rAdmin.ServeHTTP(wAdmin, httptest.NewRequest(http.MethodGet, "/api/admin/evaluations", nil))
	if wAdmin.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin, got %d: %s", wAdmin.Code, wAdmin.Body.String())
	}
	if bytes.Contains(wAdmin.Body.Bytes(), []byte("email")) {
		t.Fatalf("admin evaluation response must not contain email PII")
	}

	var dash dto.AdminEvaluationsResponse
	if err := json.Unmarshal(wAdmin.Body.Bytes(), &dash); err != nil {
		t.Fatalf("unmarshal admin dashboard: %v", err)
	}
	if len(dash.Responses) != 1 || dash.Responses[0].ParticipantCode != "P-ABCDEF12" || dash.Responses[0].Nickname != "blue_hero" {
		t.Fatalf("unexpected pseudonymized admin row: %+v", dash.Responses)
	}
	if dash.Cohort.ResponseRatePct != 100.0 {
		t.Fatalf("expected 100.0 response rate, got %v", dash.Cohort.ResponseRatePct)
	}

	// Invalid export format -> 400 EVALUATION_INVALID_EXPORT_FORMAT
	wBadFmt := httptest.NewRecorder()
	rAdmin.ServeHTTP(wBadFmt, httptest.NewRequest(http.MethodGet, "/api/admin/evaluations/export?format=xlsx", nil))
	if wBadFmt.Code != http.StatusBadRequest || !bytes.Contains(wBadFmt.Body.Bytes(), []byte("EVALUATION_INVALID_EXPORT_FORMAT")) {
		t.Fatalf("expected 400 EVALUATION_INVALID_EXPORT_FORMAT, got %d: %s", wBadFmt.Code, wBadFmt.Body.String())
	}

	// CSV export -> UTF-8 BOM + neutralized formula
	wCSV := httptest.NewRecorder()
	rAdmin.ServeHTTP(wCSV, httptest.NewRequest(http.MethodGet, "/api/admin/evaluations/export?format=csv", nil))
	if wCSV.Code != http.StatusOK {
		t.Fatalf("expected 200 CSV export, got %d", wCSV.Code)
	}
	if !bytes.HasPrefix(wCSV.Body.Bytes(), []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatalf("expected UTF-8 BOM in CSV response")
	}
	if !bytes.Contains(wCSV.Body.Bytes(), []byte("'=SUM(A1:A2)")) {
		t.Fatalf("expected formula-neutralized cell in CSV response")
	}

	// JSON export -> 200 attachment JSON
	wJSON := httptest.NewRecorder()
	rAdmin.ServeHTTP(wJSON, httptest.NewRequest(http.MethodGet, "/api/admin/evaluations/export?format=json", nil))
	if wJSON.Code != http.StatusOK {
		t.Fatalf("expected 200 JSON export, got %d", wJSON.Code)
	}
	var jsonExport dto.AdminEvaluationsExportJSONResponse
	if err := json.Unmarshal(wJSON.Body.Bytes(), &jsonExport); err != nil {
		t.Fatalf("unmarshal JSON export: %v", err)
	}
	if len(jsonExport.Responses) != 1 {
		t.Fatalf("expected 1 row in JSON export, got %d", len(jsonExport.Responses))
	}
}

func TestAdminStatistics_RBACAndSmallNGuardrails(t *testing.T) {
	store := newFakeEvalStore()
	ec := NewEvaluationControllerWithStore(store)

	// Non-admin rejected with 403
	for _, role := range []string{"learner", "specialist"} {
		rForbidden := setupEvalTestRouter(ec, "user-x", role)
		wF := httptest.NewRecorder()
		rForbidden.ServeHTTP(wF, httptest.NewRequest(http.MethodGet, "/api/admin/evaluations/statistics", nil))
		if wF.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for role %s on /statistics, got %d", role, wF.Code)
		}
	}

	// Admin on 1 respondent -> 200 with honest insufficient_n guardrails
	uid := "learner-1111-2222"
	store.userRoles[uid] = "learner"
	store.userNicknames[uid] = "stat_user"
	seedCompletedPreAndPost(store, uid, 50.0, 80.0)
	rLearner := setupEvalTestRouter(ec, uid, "learner")
	wSub := httptest.NewRecorder()
	reqSub := httptest.NewRequest(http.MethodPost, "/api/evaluation/questionnaire", bytes.NewBufferString(makeValidQuestionnaireReqBody()))
	reqSub.Header.Set("Content-Type", "application/json")
	rLearner.ServeHTTP(wSub, reqSub)
	if wSub.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", wSub.Code)
	}

	rAdmin := setupEvalTestRouter(ec, "admin-1", "admin")
	wStats := httptest.NewRecorder()
	rAdmin.ServeHTTP(wStats, httptest.NewRequest(http.MethodGet, "/api/admin/evaluations/statistics", nil))
	if wStats.Code != http.StatusOK {
		t.Fatalf("expected 200 statistics, got %d: %s", wStats.Code, wStats.Body.String())
	}

	var report dto.AdminPaperStatisticsResponse
	if err := json.Unmarshal(wStats.Body.Bytes(), &report); err != nil {
		t.Fatalf("unmarshal statistics: %v", err)
	}
	if !report.IsPilotSample || report.Cohort.QuestionnaireRespondents != 1 {
		t.Fatalf("expected pilot sample n=1, got pilot=%v n=%d", report.IsPilotSample, report.Cohort.QuestionnaireRespondents)
	}
	if report.LearningGain.PairedTTest.Status != "insufficient_n" || report.LearningGain.Wilcoxon.Status != "insufficient_n" {
		t.Fatalf("expected insufficient_n gates on n=1: t=%s w=%s", report.LearningGain.PairedTTest.Status, report.LearningGain.Wilcoxon.Status)
	}
	if len(report.ParticipantPoints) != 1 || report.ParticipantPoints[0].ParticipantCode != "P-LEARNER1" {
		t.Fatalf("expected 1 pseudonymized ParticipantPoint P-LEARNER1, got %+v", report.ParticipantPoints)
	}
	if report.ParticipantPoints[0].PreScore != 50.0 || report.ParticipantPoints[0].PostScore != 80.0 || report.ParticipantPoints[0].ImprovementPct != 30.0 {
		t.Fatalf("unexpected ParticipantPoint scores: %+v", report.ParticipantPoints[0])
	}
}

func TestGetAdminPaperStatisticsEmptyCohort(t *testing.T) {
	store := newFakeEvalStore()
	ec := NewEvaluationControllerWithStore(store)
	rAdmin := setupEvalTestRouter(ec, "admin-1", "admin")

	w := httptest.NewRecorder()
	rAdmin.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/admin/evaluations/statistics", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on empty cohort, got %d: %s", w.Code, w.Body.String())
	}

	raw := w.Body.String()
	if bytes.Contains([]byte(raw), []byte("NaN")) || bytes.Contains([]byte(raw), []byte("Inf")) {
		t.Fatalf("unexpected non-finite float marker in response: %s", raw)
	}

	var report dto.AdminPaperStatisticsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &report); err != nil {
		t.Fatalf("unmarshal empty statistics: %v", err)
	}
	if len(report.ParticipantPoints) != 0 {
		t.Fatalf("expected 0 ParticipantPoints on empty cohort, got %d", len(report.ParticipantPoints))
	}
	if report.LearningGain.PairedTTest.Status != "insufficient_n" {
		t.Fatalf("expected insufficient_n gate on n=0, got %s", report.LearningGain.PairedTTest.Status)
	}
}

func TestSwaggerSpecDefinesParticipantPlotPoints(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "docs", "swagger.json"))
	if err != nil {
		t.Fatalf("read swagger.json: %v", err)
	}

	var spec struct {
		Definitions map[string]struct {
			Properties map[string]json.RawMessage `json:"properties"`
		} `json:"definitions"`
	}
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("parse swagger.json: %v", err)
	}

	statsDef, ok := spec.Definitions["dto.AdminPaperStatisticsResponse"]
	if !ok {
		t.Fatalf("missing dto.AdminPaperStatisticsResponse in swagger.json")
	}
	if _, ok := statsDef.Properties["participant_points"]; !ok {
		t.Fatalf("expected participant_points property in dto.AdminPaperStatisticsResponse")
	}

	pointDef, ok := spec.Definitions["dto.ParticipantPlotPointDTO"]
	if !ok {
		t.Fatalf("missing dto.ParticipantPlotPointDTO in swagger.json")
	}
	for _, field := range []string{"participant_code", "pre_score", "post_score", "improvement_pct", "csuq_overall", "nasa_tlx_overall"} {
		if _, ok := pointDef.Properties[field]; !ok {
			t.Fatalf("expected %s property in dto.ParticipantPlotPointDTO", field)
		}
	}
}


