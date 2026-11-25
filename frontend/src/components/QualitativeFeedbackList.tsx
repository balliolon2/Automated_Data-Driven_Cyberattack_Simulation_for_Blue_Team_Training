interface QualitativeFeedbackValues {
  most_impressive: string;
  areas_to_improve: string;
  additional_comments: string;
}

interface QualitativeFeedbackListProps {
  feedback: QualitativeFeedbackValues;
  labels: {
    mostImpressive: string;
    areasToImprove: string;
    additionalComments: string;
    emptyText: string;
  };
  hasNotes?: boolean;
  cardClassName?: string;
}

/** Shared stacked renderer for optional open-ended learner impression prompts. */
export default function QualitativeFeedbackList({
  feedback,
  labels,
  hasNotes = true,
  cardClassName = "rounded-md border border-graphite-800 bg-graphite-950/60 p-4 space-y-1.5",
}: QualitativeFeedbackListProps) {
  const entries = [
    { label: labels.mostImpressive, value: feedback.most_impressive },
    { label: labels.areasToImprove, value: feedback.areas_to_improve },
    { label: labels.additionalComments, value: feedback.additional_comments },
  ].filter((item) => item.value.trim().length > 0);

  if (!hasNotes || entries.length === 0) {
    return <p className="text-sm text-graphite-400">{labels.emptyText}</p>;
  }

  return (
    <div className="space-y-3">
      {entries.map((item) => (
        <div key={item.label} className={cardClassName}>
          <div className="text-xs font-semibold text-emerald-400">{item.label}</div>
          <p className="text-base text-graphite-200 leading-relaxed whitespace-pre-line">{item.value}</p>
        </div>
      ))}
    </div>
  );
}
