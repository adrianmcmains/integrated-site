package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/adrianmcmains/integrated-site/models"
)

type PostRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *models.Post) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert post
	query := `
		INSERT INTO blog.posts (title, slug, content, excerpt, featured_image, author_id, status, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(ctx, query,
		post.Title,
		post.Slug,
		post.Content,
		post.Excerpt,
		post.FeaturedImage,
		post.AuthorID,
		post.Status,
		post.PublishedAt,
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return err
	}

	// Insert categories
	if len(post.Categories) > 0 {
		for _, category := range post.Categories {
			_, err = tx.Exec(ctx, `
				INSERT INTO blog.post_categories (post_id, category_id)
				VALUES ($1, $2)
			`, post.ID, category.ID)
			if err != nil {
				return err
			}
		}
	}

	// Insert tags
	if len(post.Tags) > 0 {
		for _, tag := range post.Tags {
			_, err = tx.Exec(ctx, `
				INSERT INTO blog.post_tags (post_id, tag_id)
				VALUES ($1, $2)
			`, post.ID, tag.ID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func (r *PostRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	query := `
		SELECT p.id, p.title, p.slug, p.content, p.excerpt, p.featured_image, 
			   p.author_id, p.status, p.published_at, p.created_at, p.updated_at,
			   a.id, a.user_id, a.bio, a.social_media, a.created_at, a.updated_at,
			   u.id, u.email, u.full_name, u.role, u.avatar_url, u.created_at, u.updated_at
		FROM blog.posts p
		LEFT JOIN blog.authors a ON p.author_id = a.id
		LEFT JOIN auth.users u ON a.user_id = u.id
		WHERE p.id = $1
	`

	var post models.Post
	var author models.Author
	var user models.User
	var socialMediaJSON []byte
	var publishedAt *time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&post.ID, &post.Title, &post.Slug, &post.Content, &post.Excerpt, &post.FeaturedImage,
		&post.AuthorID, &post.Status, &publishedAt, &post.CreatedAt, &post.UpdatedAt,
		&author.ID, &author.UserID, &author.Bio, &socialMediaJSON, &author.CreatedAt, &author.UpdatedAt,
		&user.ID, &user.Email, &user.FullName, &user.Role, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	post.PublishedAt = publishedAt
	author.User = &user
	post.Author = &author

	// Get categories
	categoriesQuery := `
		SELECT c.id, c.name, c.slug, c.description, c.created_at, c.updated_at
		FROM blog.categories c
		JOIN blog.post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = $1
	`

	categoryRows, err := r.db.Query(ctx, categoriesQuery, post.ID)
	if err != nil {
		return nil, err
	}
	defer categoryRows.Close()

	post.Categories = []*models.Category{}
	for categoryRows.Next() {
		var category models.Category
		if err := categoryRows.Scan(
			&category.ID, &category.Name, &category.Slug, &category.Description,
			&category.CreatedAt, &category.UpdatedAt,
		); err != nil {
			return nil, err
		}
		post.Categories = append(post.Categories, &category)
	}

	// Get tags
	tagsQuery := `
		SELECT t.id, t.name, t.slug, t.created_at, t.updated_at
		FROM blog.tags t
		JOIN blog.post_tags pt ON t.id = pt.tag_id
		WHERE pt.post_id = $1
	`

	tagRows, err := r.db.Query(ctx, tagsQuery, post.ID)
	if err != nil {
		return nil, err
	}
	defer tagRows.Close()

	post.Tags = []*models.Tag{}
	for tagRows.Next() {
		var tag models.Tag
		if err := tagRows.Scan(
			&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt, &tag.UpdatedAt,
		); err != nil {
			return nil, err
		}
		post.Tags = append(post.Tags, &tag)
	}

	return &post, nil
}

func (r *PostRepository) List(ctx context.Context, limit, offset int, status string) ([]*models.Post, error) {
	query := `
		SELECT p.id, p.title, p.slug, p.excerpt, p.featured_image, 
			   p.author_id, p.status, p.published_at, p.created_at, p.updated_at
		FROM blog.posts p
	`

	args := []interface{}{}
	if status != "" {
		query += " WHERE p.status = $1"
		args = append(args, status)
	}

	query += " ORDER BY p.published_at DESC, p.created_at DESC LIMIT $" + 
		 		string(len(args) + 1) + " OFFSET $" + string(len(args) + 2)
	
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		var publishedAt *time.Time

		if err := rows.Scan(
			&post.ID, &post.Title, &post.Slug, &post.Excerpt, &post.FeaturedImage,
			&post.AuthorID, &post.Status, &publishedAt, &post.CreatedAt, &post.UpdatedAt,
		); err != nil {
			return nil, err
		}

		post.PublishedAt = publishedAt
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *PostRepository) Update(ctx context.Context, post *models.Post) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update post
	query := `
		UPDATE blog.posts
		SET title = $1, slug = $2, content = $3, excerpt = $4, 
			featured_image = $5, status = $6, published_at = $7
		WHERE id = $8
		RETURNING updated_at
	`

	err = tx.QueryRow(ctx, query,
		post.Title,
		post.Slug,
		post.Content,
		post.Excerpt,
		post.FeaturedImage,
		post.Status,
		post.PublishedAt,
		post.ID,
	).Scan(&post.UpdatedAt)
	if err != nil {
		return err
	}

	// Delete old categories
	_, err = tx.Exec(ctx, "DELETE FROM blog.post_categories WHERE post_id = $1", post.ID)
	if err != nil {
		return err
	}

	// Insert new categories
	if len(post.Categories) > 0 {
		for _, category := range post.Categories {
			_, err = tx.Exec(ctx, `
				INSERT INTO blog.post_categories (post_id, category_id)
				VALUES ($1, $2)
			`, post.ID, category.ID)
			if err != nil {
				return err
			}
		}
	}

	// Delete old tags
	_, err = tx.Exec(ctx, "DELETE FROM blog.post_tags WHERE post_id = $1", post.ID)
	if err != nil {
		return err
	}

	// Insert new tags
	if len(post.Tags) > 0 {
		for _, tag := range post.Tags {
			_, err = tx.Exec(ctx, `
				INSERT INTO blog.post_tags (post_id, tag_id)
				VALUES ($1, $2)
			`, post.ID, tag.ID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func (r *PostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM blog.posts WHERE id = $1", id)
	return err
}

func (r *PostRepository) Count(ctx context.Context, status string) (int, error) {
	query := `SELECT COUNT(*) FROM blog.posts`
	args := []interface{}{}

	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}

	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}
SELECT t.id, t.name, t.slug, t.created_at, t.updated_at
		FROM blog.tags t
		JOIN blog.post_tags pt ON t.id = pt.tag_id
		WHERE pt.post_id = $1
	`

	tagRows, err := r.db.Query(ctx, tagsQuery, post.ID)
	if err != nil {
		return nil, err
	}
	defer tagRows.Close()

	post.Tags = []*models.Tag{}
	for tagRows.Next() {
		var tag models.Tag
		if err := tagRows.Scan(
			&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt, &tag.UpdatedAt,
		); err != nil {
			return nil, err
		}
		post.Tags = append(post.Tags, &tag)
	}

	return &post, nil
}

func (r *PostRepository) GetBySlug(ctx context.Context, slug string) (*models.Post, error) {
	query := `
		SELECT p.id, p.title, p.slug, p.content, p.excerpt, p.featured_image, 
			   p.author_id, p.status, p.published_at, p.created_at, p.updated_at,
			   a.id, a.user_id, a.bio, a.social_media, a.created_at, a.updated_at,
			   u.id, u.email, u.full_name, u.role, u.avatar_url, u.created_at, u.updated_at
		FROM blog.posts p
		LEFT JOIN blog.authors a ON p.author_id = a.id
		LEFT JOIN auth.users u ON a.user_id = u.id
		WHERE p.slug = $1
	`

	var post models.Post
	var author models.Author
	var user models.User
	var socialMediaJSON []byte
	var publishedAt *time.Time

	err := r.db.QueryRow(ctx, query, slug).Scan(
		&post.ID, &post.Title, &post.Slug, &post.Content, &post.Excerpt, &post.FeaturedImage,
		&post.AuthorID, &post.Status, &publishedAt, &post.CreatedAt, &post.UpdatedAt,
		&author.ID, &author.UserID, &author.Bio, &socialMediaJSON, &author.CreatedAt, &author.UpdatedAt,
		&user.ID, &user.Email, &user.FullName, &user.Role, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	post.PublishedAt = publishedAt
	author.User = &user
	post.Author = &author

	// Get categories
	categoriesQuery := `
		SELECT c.id, c.name, c.slug, c.description, c.created_at, c.updated_at
		FROM blog.categories c
		JOIN blog.post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = $1
	`

	categoryRows, err := r.db.Query(ctx, categoriesQuery, post.ID)
	if err != nil {
		return nil, err
	}
	defer categoryRows.Close()

	post.Categories = []*models.Category{}
	for categoryRows.Next() {
		var category models.Category
		if err := categoryRows.Scan(
			&category.ID, &category.Name, &category.Slug, &category.Description,
			&category.CreatedAt, &category.UpdatedAt,
		); err != nil {
			return nil, err
		}
		post.Categories = append(post.Categories, &category)
	}

	// Get tags
	tagsQuery := `