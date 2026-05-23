## 1. Update Go Model

- [x] 1.1 Open `models/user.go`
- [x] 1.2 Remove `UpdatedAt` and `DeletedAt` fields
- [x] 1.3 Add `LastLogin *time.Time` and `IsActive bool` fields with appropriate GORM tags
- [x] 1.4 Restart Go server and verify registration works from React frontend
