package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	FullName     string     `json:"full_name"`
	Role         string     `json:"role"`
	AvatarURL    string     `json:"avatar_url,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Blog models
type Author struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	Bio        string     `json:"bio,omitempty"`
	SocialMedia map[string]string `json:"social_media,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	User       *User      `json:"user,omitempty"`
}

type Category struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Tag struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Post struct {
	ID            uuid.UUID   `json:"id"`
	Title         string      `json:"title"`
	Slug          string      `json:"slug"`
	Content       string      `json:"content"`
	Excerpt       string      `json:"excerpt,omitempty"`
	FeaturedImage string      `json:"featured_image,omitempty"`
	AuthorID      uuid.UUID   `json:"author_id"`
	Status        string      `json:"status"`
	PublishedAt   *time.Time  `json:"published_at,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	Author        *Author     `json:"author,omitempty"`
	Categories    []*Category `json:"categories,omitempty"`
	Tags          []*Tag      `json:"tags,omitempty"`
	Comments      []*Comment  `json:"comments,omitempty"`
}

type Comment struct {
	ID        uuid.UUID  `json:"id"`
	PostID    uuid.UUID  `json:"post_id"`
	UserID    uuid.UUID  `json:"user_id"`
	Content   string     `json:"content"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	User      *User      `json:"user,omitempty"`
	Parent    *Comment   `json:"parent,omitempty"`
	Replies   []*Comment `json:"replies,omitempty"`
}

// E-commerce models
type ProductCategory struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description,omitempty"`
	Image       string     `json:"image,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Products    []*Product `json:"products,omitempty"`
}

type Product struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Slug        string           `json:"slug"`
	Description string           `json:"description"`
	Price       float64          `json:"price"`
	SalePrice   *float64         `json:"sale_price,omitempty"`
	SKU         string           `json:"sku"`
	Stock       int              `json:"stock"`
	IsFeatured  bool             `json:"is_featured"`
	Images      []string         `json:"images,omitempty"`
	CategoryID  uuid.UUID        `json:"category_id"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Category    *ProductCategory `json:"category,omitempty"`
	Attributes  []*ProductAttribute `json:"attributes,omitempty"`
}

type ProductAttribute struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Customer struct {
	ID             uuid.UUID         `json:"id"`
	UserID         uuid.UUID         `json:"user_id"`
	ShippingAddress map[string]string `json:"shipping_address,omitempty"`
	BillingAddress  map[string]string `json:"billing_address,omitempty"`
	Phone          string            `json:"phone,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	User           *User             `json:"user,omitempty"`
	Orders         []*Order          `json:"orders,omitempty"`
}

type Order struct {
	ID              uuid.UUID         `json:"id"`
	CustomerID      uuid.UUID         `json:"customer_id"`
	Status          string            `json:"status"`
	TotalAmount     float64           `json:"total_amount"`
	ShippingAddress map[string]string `json:"shipping_address"`
	BillingAddress  map[string]string `json:"billing_address"`
	PaymentMethod   string            `json:"payment_method"`
	PaymentStatus   string            `json:"payment_status"`
	TrackingNumber  string            `json:"tracking_number,omitempty"`
	Notes           string            `json:"notes,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	Customer        *Customer         `json:"customer,omitempty"`
	Items           []*OrderItem      `json:"items,omitempty"`
	Payment         *Payment          `json:"payment,omitempty"`
}

type OrderItem struct {
	ID        uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Product   *Product  `json:"product,omitempty"`
}

type Payment struct {
	ID             uuid.UUID         `json:"id"`
	OrderID        uuid.UUID         `json:"order_id"`
	Amount         float64           `json:"amount"`
	PaymentMethod  string            `json:"payment_method"`
	PaymentID      string            `json:"payment_id,omitempty"`
	Status         string            `json:"status"`
	TransactionData map[string]interface{} `json:"transaction_data,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// CMS models
type SiteSetting struct {
	ID        uuid.UUID              `json:"id"`
	Key       string                 `json:"key"`
	Value     map[string]interface{} `json:"value"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type Page struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Slug           string    `json:"slug"`
	Content        string    `json:"content"`
	MetaTitle      string    `json:"meta_title,omitempty"`
	MetaDescription string    `json:"meta_description,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Auth models
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin customer contributor"`
}

type TokenResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         User      `json:"user"`
}