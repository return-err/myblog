package repository

import (
	"testing"

	"github.com/return-err/myblog/internal/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePost(t *testing.T) {
	repo := NewPostRepository(testDB)
	t.Run("success", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := initAdmin(testDB)
		require.NoError(t, err)
		err = repo.CreatePost(1, "test", "")
		require.NoError(t, err)

		var title string
		err = testDB.QueryRow("SELECT title FROM posts WHERE user_id = 1").Scan(&title)
		require.NoError(t, err)
		assert.Equal(t, "test", title)
	})

	t.Run("invalid user", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := repo.CreatePost(1, "test", "")
		require.Error(t, err)
	})
}

func TestGetTitleByUserID(t *testing.T) {
	postRepo := NewPostRepository(testDB)
	authRepo := NewAuthRepository(testDB)

	resetTable(t)
	defer resetTable(t)

	authRepo.CreateUserWithProfile("User1", "pass")
	authRepo.CreateUserWithProfile("User2", "pass")
	postRepo.CreatePost(1, "User1 - Post 1", "")
	postRepo.CreatePost(1, "User1 - Post 2", "")
	postRepo.CreatePost(1, "User1 - Post 3", "")
	postRepo.CreatePost(2, "User2 - Post 1", "")

	tests := []struct {
		name           string
		userID         int64
		page           int
		pageSize       int
		wantPageCount  int
		wantTotal      int64
		wantFirstTitle string
	}{
		{
			name:           "User1 Page 1",
			userID:         1,
			page:           1,
			pageSize:       2,
			wantPageCount:  2,
			wantTotal:      3,
			wantFirstTitle: "User1 - Post 3",
		},
		{
			name:           "User1 Page 2",
			userID:         1,
			page:           2,
			pageSize:       2,
			wantPageCount:  1,
			wantTotal:      3,
			wantFirstTitle: "User1 - Post 1",
		},
		{
			name:          "invalid user",
			userID:        10,
			page:          1,
			pageSize:      10,
			wantPageCount: 0,
			wantTotal:     0,
		},
		{
			name:          "invalid page",
			userID:        2,
			page:          2,
			pageSize:      2,
			wantPageCount: 0,
			wantTotal:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, total, err := postRepo.GetTitleByUserID(tt.userID, tt.page, tt.pageSize)

			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, total)
			assert.Len(t, posts, tt.wantPageCount)

			if tt.wantPageCount > 0 {
				assert.Equal(t, tt.wantFirstTitle, posts[0].Title)
			}
		})
	}
}

func TestGetPostDetail(t *testing.T) {
	resetTable(t)
	defer resetTable(t)
	postRepo := NewPostRepository(testDB)
	err := initAdmin(testDB) // デフォルト DisplayName: ユーザー1
	require.NoError(t, err)
	err = postRepo.CreatePost(1, "test", "")
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		post, err := postRepo.GetPostDetail(1)
		require.NoError(t, err)
		assert.Equal(t, "test", post.Title)
		assert.Equal(t, "", post.Content)
	})

	t.Run("invalid post", func(t *testing.T) {
		_, err := postRepo.GetPostDetail(2)
		require.Error(t, err)
	})
}

func TestDeletePost(t *testing.T) {
	postRepo := NewPostRepository(testDB)

	t.Run("success", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := initAdmin(testDB)
		require.NoError(t, err)
		err = postRepo.CreatePost(1, "test", "")
		require.NoError(t, err)
		err = postRepo.DeletePost(1)
		require.NoError(t, err)

		var count int
		err = testDB.QueryRow("SELECT COUNT(*) FROM posts WHERE id = $1", 1).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("invalid post", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := postRepo.DeletePost(2)
		require.Error(t, err)
		assert.Equal(t, apperror.ErrPostNotFound, err)
	})
}

func TestEditPost(t *testing.T) {
	postRepo := NewPostRepository(testDB)

	t.Run("success", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := initAdmin(testDB)
		require.NoError(t, err)
		err = postRepo.CreatePost(1, "編集前", "編集前コンテンツ")
		require.NoError(t, err)

		err = postRepo.EditPost(1, "編集後", "編集後コンテンツ")
		require.NoError(t, err)
		var title, content string
		err = testDB.QueryRow(`
			SELECT title, content
			FROM posts
			WHERE id = $1
		`, 1).Scan(&title, &content)
		require.NoError(t, err)
		assert.Equal(t, "編集後", title)
		assert.Equal(t, "編集後コンテンツ", content)
	})

	t.Run("invalid post", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := postRepo.EditPost(1, "編集後", "編集後コンテンツ")
		require.Error(t, err)
		assert.Equal(t, apperror.ErrPostNotFound, err)
	})
}

func TestGetAllPosts(t *testing.T) {
	resetTable(t)
	defer resetTable(t)

	err := initAdmin(testDB)
	require.NoError(t, err)
	postRepo := NewPostRepository(testDB)

	titles := []string{"post1", "post2", "post3", "post4", "post5"}
	for _, title := range titles {
		err := postRepo.CreatePost(1, title, "content")
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		page           int
		pageSize       int
		wantCount      int64
		wantLen        int
		wantFirstTitle string
	}{
		{
			name:           "Page 1",
			page:           1,
			pageSize:       2,
			wantCount:      5,
			wantLen:        2,
			wantFirstTitle: "post5",
		},
		{
			name:           "Page 2",
			page:           2,
			pageSize:       2,
			wantCount:      5,
			wantLen:        2,
			wantFirstTitle: "post3",
		},
		{
			name:           "Page 3",
			page:           3,
			pageSize:       2,
			wantCount:      5,
			wantLen:        1,
			wantFirstTitle: "post1",
		},
		{
			name:      "Invalid page",
			page:      4,
			pageSize:  2,
			wantCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, count, err := postRepo.GetAllPosts(tt.page, tt.pageSize)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
			require.Len(t, posts, tt.wantLen)

			if tt.wantLen > 0 {
				assert.Equal(t, tt.wantFirstTitle, posts[0].Title)
			}
		})
	}
}

func TestSearchPost(t *testing.T) {
	resetTable(t)
	defer resetTable(t)
	initAdmin(testDB)
	postRepo := NewPostRepository(testDB)

	titles := []string{"post1", "post2", "post3", "post4", "post5"}
	for _, title := range titles {
		err := postRepo.CreatePost(1, title, "content")
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		keyword string
		wantLen int
	}{
		{
			name:    "match mulpitle",
			keyword: "post",
			wantLen: 5,
		},
		{
			name:    "match single",
			keyword: "post1",
			wantLen: 1,
		},
		{
			name:    "missing post",
			keyword: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, err := postRepo.SearchPost(tt.keyword)
			require.NoError(t, err)
			require.Len(t, posts, tt.wantLen)
		})
	}
}
