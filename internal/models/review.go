package models

import (
	"time"

	"github.com/google/uuid"
)

// ReviewStatus represents the status of a review
type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"
	ReviewStatusApproved ReviewStatus = "approved"
	ReviewStatusRejected ReviewStatus = "rejected"
)

// Review represents a product review
type Review struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	ProductID uuid.UUID    `json:"product_id" db:"product_id"`
	UserID    uuid.UUID    `json:"user_id" db:"user_id"`
	OrderID   *uuid.UUID   `json:"order_id,omitempty" db:"order_id"`
	Rating    int          `json:"rating" db:"rating"`
	Title     string       `json:"title" db:"title"`
	Comment   string       `json:"comment" db:"comment"`
	Status    ReviewStatus `json:"status" db:"status"`
	IsVerifiedPurchase bool `json:"is_verified_purchase" db:"is_verified_purchase"`
	HelpfulCount int       `json:"helpful_count" db:"helpful_count"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	User      *User        `json:"user,omitempty"`
	Product   *Product     `json:"product,omitempty"`
}

// CreateReviewRequest represents the request for creating a review
type CreateReviewRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	OrderID   *uuid.UUID `json:"order_id,omitempty"`
	Rating    int       `json:"rating" binding:"required,min=1,max=5"`
	Title     string    `json:"title" binding:"required,min=3,max=100"`
	Comment   string    `json:"comment" binding:"required,min=10,max=1000"`
}

// UpdateReviewRequest represents the request for updating a review
type UpdateReviewRequest struct {
	Rating  *int    `json:"rating,omitempty" binding:"omitempty,min=1,max=5"`
	Title   *string `json:"title,omitempty" binding:"omitempty,min=3,max=100"`
	Comment *string `json:"comment,omitempty" binding:"omitempty,min=10,max=1000"`
}

// ReviewSearchRequest represents search parameters for reviews
type ReviewSearchRequest struct {
	ProductID *uuid.UUID    `form:"product_id"`
	UserID    *uuid.UUID    `form:"user_id"`
	Rating    *int          `form:"rating"`
	Status    *ReviewStatus `form:"status"`
	VerifiedOnly bool       `form:"verified_only"`
	Page      int           `form:"page" binding:"min=1"`
	Size      int           `form:"size" binding:"min=1,max=100"`
	SortBy    string        `form:"sort_by"`
	SortOrder string        `form:"sort_order"`
}

// ReviewSearchResponse represents search results for reviews
type ReviewSearchResponse struct {
	Reviews []Review `json:"reviews"`
	Total   int64    `json:"total"`
	Page    int      `json:"page"`
	Size    int      `json:"size"`
	Summary ReviewSummary `json:"summary"`
}

// ReviewSummary represents aggregated review data
type ReviewSummary struct {
	AverageRating float64            `json:"average_rating"`
	TotalReviews  int64              `json:"total_reviews"`
	RatingDistribution map[int]int64 `json:"rating_distribution"`
	VerifiedPurchaseCount int64       `json:"verified_purchase_count"`
}

// ProductReviewSummary represents review summary for a specific product
type ProductReviewSummary struct {
	ProductID      uuid.UUID `json:"product_id"`
	AverageRating  float64   `json:"average_rating"`
	TotalReviews   int64     `json:"total_reviews"`
	RatingCounts   map[int]int64 `json:"rating_counts"`
	VerifiedCount  int64     `json:"verified_count"`
	RecentReviews  []Review  `json:"recent_reviews"`
}

// ReviewModeration represents review moderation actions
type ReviewModeration struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	ReviewID  uuid.UUID    `json:"review_id" db:"review_id"`
	ModeratorID uuid.UUID  `json:"moderator_id" db:"moderator_id"`
	Action    ReviewStatus `json:"action" db:"action"`
	Reason    string       `json:"reason" db:"reason"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	Moderator *User        `json:"moderator,omitempty"`
}

// ReviewModerationRequest represents the request for moderating a review
type ReviewModerationRequest struct {
	Action ReviewStatus `json:"action" binding:"required,oneof=approved rejected"`
	Reason string       `json:"reason" binding:"required,min=3,max=500"`
}

// ReviewAnalytics represents review analytics data
type ReviewAnalytics struct {
	TotalReviews       int64             `json:"total_reviews"`
	AverageRating      float64           `json:"average_rating"`
	ReviewsByRating    map[int]int64     `json:"reviews_by_rating"`
	ReviewsByStatus    map[ReviewStatus]int64 `json:"reviews_by_status"`
	VerifiedPercentage float64           `json:"verified_percentage"`
	MonthlyReviews     map[string]int64  `json:"monthly_reviews"`
	TopReviewedProducts []ProductReviewSummary `json:"top_reviewed_products"`
}

// IsPositive checks if the review is positive (4-5 stars)
func (r *Review) IsPositive() bool {
	return r.Rating >= 4
}

// IsNegative checks if the review is negative (1-2 stars)
func (r *Review) IsNegative() bool {
	return r.Rating <= 2
}

// CanBeModerated checks if the review can be moderated
func (r *Review) CanBeModerated() bool {
	return r.Status == ReviewStatusPending
}

// CalculateAverageRating calculates the average rating from rating distribution
func (rs *ReviewSummary) CalculateAverageRating() {
	if rs.TotalReviews == 0 {
		rs.AverageRating = 0
		return
	}
	
	total := 0.0
	count := 0.0
	
	for rating, reviewCount := range rs.RatingDistribution {
		total += float64(rating) * float64(reviewCount)
		count += float64(reviewCount)
	}
	
	if count > 0 {
		rs.AverageRating = total / count
	}
}