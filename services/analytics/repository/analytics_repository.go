package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type MentorStat struct {
	MentorID          uint
	QuestionsAnswered int
	AcceptedSolutions int
	TotalUpvotes      int
}

type StudentStat struct {
	StudentID         uint
	QuestionsAsked    int
	QuestionsAnswered int
}

type TopicWeakness struct {
	Subject        string
	Topic          string
	QuestionCount  int
	ResolutionRate float64
}

type PlatformMetrics struct {
	ActiveStudents   int
	ActiveMentors    int
	TotalQuestions   int
	PendingQuestions int
	AvgResponseHours float64
	ResolutionRate   float64
}

type Repository interface {
	GetLeaderboard(ctx context.Context, since time.Time, limit int) ([]MentorStat, error)
	GetStudentStats(ctx context.Context, studentID uint, since time.Time) (*StudentStat, error)
	GetMentorStats(ctx context.Context, mentorID uint, since time.Time) (*MentorStat, error)
	GetPlatformMetrics(ctx context.Context, since time.Time) (*PlatformMetrics, error)
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetLeaderboard(ctx context.Context, since time.Time, limit int) ([]MentorStat, error) {
	type row struct {
		MentorID          uint
		QuestionsAnswered int
		AcceptedSolutions int
		TotalUpvotes      int
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			s.mentor_id,
			COUNT(DISTINCT q.id) AS questions_answered,
			COUNT(DISTINCT CASE WHEN s.is_accepted THEN s.id END) AS accepted_solutions,
			COALESCE(SUM(s.upvotes), 0)::int AS total_upvotes
		FROM solutions s
		JOIN questions q ON q.id = s.question_id
		WHERE s.created_at >= ?
		GROUP BY s.mentor_id
		ORDER BY total_upvotes DESC, questions_answered DESC
		LIMIT ?
	`, since, limit).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("analytics_repo_leaderboard: %w", err)
	}
	result := make([]MentorStat, len(rows))
	for i, row := range rows {
		result[i] = MentorStat{
			MentorID:          row.MentorID,
			QuestionsAnswered: row.QuestionsAnswered,
			AcceptedSolutions: row.AcceptedSolutions,
			TotalUpvotes:      row.TotalUpvotes,
		}
	}
	return result, nil
}

func (r *repository) GetStudentStats(ctx context.Context, studentID uint, since time.Time) (*StudentStat, error) {
	var stat StudentStat
	stat.StudentID = studentID

	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)::int AS questions_asked
		FROM questions
		WHERE student_id = ? AND created_at >= ?
	`, studentID, since).Scan(&stat).Error
	if err != nil {
		return nil, fmt.Errorf("analytics_repo_student_stats: %w", err)
	}

	err = r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT q.id)::int AS questions_answered
		FROM questions q
		JOIN solutions s ON s.question_id = q.id
		WHERE q.student_id = ? AND s.is_accepted = TRUE AND q.created_at >= ?
	`, studentID, since).Scan(&stat.QuestionsAnswered).Error
	if err != nil {
		return nil, fmt.Errorf("analytics_repo_student_answered: %w", err)
	}

	return &stat, nil
}

func (r *repository) GetMentorStats(ctx context.Context, mentorID uint, since time.Time) (*MentorStat, error) {
	var stat MentorStat
	stat.MentorID = mentorID

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(DISTINCT s.id)::int AS questions_answered,
			COUNT(DISTINCT CASE WHEN s.is_accepted THEN s.id END)::int AS accepted_solutions,
			COALESCE(SUM(s.upvotes), 0)::int AS total_upvotes
		FROM solutions s
		WHERE s.mentor_id = ? AND s.created_at >= ?
	`, mentorID, since).Scan(&stat).Error
	if err != nil {
		return nil, fmt.Errorf("analytics_repo_mentor_stats: %w", err)
	}
	return &stat, nil
}

func (r *repository) GetPlatformMetrics(ctx context.Context, since time.Time) (*PlatformMetrics, error) {
	var m PlatformMetrics

	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT student_id)::int FROM questions WHERE created_at >= ?
	`, since).Scan(&m.ActiveStudents)

	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT mentor_id)::int FROM solutions WHERE created_at >= ?
	`, since).Scan(&m.ActiveMentors)

	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)::int FROM questions WHERE created_at >= ?
	`, since).Scan(&m.TotalQuestions)

	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)::int FROM questions WHERE status = 'open' AND created_at >= ?
	`, since).Scan(&m.PendingQuestions)

	r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (q.updated_at - q.created_at))) / 3600, 0)
		FROM questions q
		WHERE q.status IN ('answered', 'assigned') AND q.created_at >= ?
	`, since).Scan(&m.AvgResponseHours)

	r.db.WithContext(ctx).Raw(`
		SELECT CASE
			WHEN COUNT(*) = 0 THEN 0
			ELSE COUNT(DISTINCT CASE WHEN q.status = 'answered' THEN q.id END)::float / COUNT(*)::float
		END
		FROM questions q
		WHERE q.created_at >= ?
	`, since).Scan(&m.ResolutionRate)

	return &m, nil
}
