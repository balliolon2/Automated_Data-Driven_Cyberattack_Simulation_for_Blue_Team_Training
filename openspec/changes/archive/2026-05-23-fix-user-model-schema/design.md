## Context

The Go server crashes during user registration because GORM tries to insert values into `updated_at` and `deleted_at` columns which do not exist in the PostgreSQL `users` table created by `init.sql`. 

## Goals / Non-Goals

**Goals:**
- Update `models.User` struct to match the `init.sql` schema.
- Ensure GORM can successfully insert a new user without referencing non-existent columns.

**Non-Goals:**
- Modify the database schema (`init.sql`) to match GORM defaults. We want the database to be the source of truth.

## Decisions

- **Model Update:** We will manually map the Go struct fields to exactly what exists in the database.
  - Remove `UpdatedAt time.Time`
  - Remove `DeletedAt gorm.DeletedAt`
  - Add `LastLogin *time.Time`
  - Add `IsActive bool` with `gorm:"default:true"`
- **GORM `Create` Behavior:** When inserting, GORM will only include fields defined in the struct, effectively bypassing the missing columns error.

## Risks / Trade-offs

- None. This aligns the application logic with the defined database schema.
