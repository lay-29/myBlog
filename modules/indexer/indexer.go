package indexer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/blevesearch/bleve/v2"
	"myblog/models"
)

type ArticleIndexer struct {
	index bleve.Index
}

type ArticleDocument struct {
	ID         int64  `json:"id"`
	TeamID     int64  `json:"team_id"`
	SpaceID    int64  `json:"space_id"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	Content    string `json:"content"`
	Ngrams     string `json:"ngrams"`
	Visibility string `json:"visibility"`
}

type SearchHit struct {
	ID    int64
	Score float64
}

func Open(path string) (*ArticleIndexer, error) {
	mapping := bleve.NewIndexMapping()
	mapping.DefaultAnalyzer = "standard"

	var (
		idx bleve.Index
		err error
	)
	if path == "" {
		idx, err = bleve.NewMemOnly(mapping)
	} else if _, statErr := os.Stat(path); statErr == nil {
		idx, err = bleve.Open(path)
	} else if os.IsNotExist(statErr) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		idx, err = bleve.New(path, mapping)
	} else {
		return nil, statErr
	}
	if err != nil {
		return nil, err
	}
	return &ArticleIndexer{index: idx}, nil
}

func (idx *ArticleIndexer) Close() error {
	if idx == nil || idx.index == nil {
		return nil
	}
	return idx.index.Close()
}

func (idx *ArticleIndexer) IndexArticle(article *models.Article) error {
	if idx == nil || idx.index == nil || article == nil {
		return nil
	}
	doc := ArticleDocument{
		ID:         article.ID,
		TeamID:     article.TeamID,
		SpaceID:    article.SpaceID,
		Title:      article.Title,
		Summary:    article.Summary,
		Content:    stripHTML(article.ContentHTML),
		Ngrams:     Ngrams(article.Title + " " + article.Summary + " " + article.ContentMarkdown),
		Visibility: string(article.Visibility),
	}
	return idx.index.Index(docID(article.ID), doc)
}

func (idx *ArticleIndexer) DeleteArticle(id int64) error {
	if idx == nil || idx.index == nil {
		return nil
	}
	return idx.index.Delete(docID(id))
}

func (idx *ArticleIndexer) Search(ctx context.Context, query string, limit int) ([]SearchHit, error) {
	if idx == nil || idx.index == nil {
		return nil, nil
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	textQuery := bleve.NewMatchQuery(query)
	ngramQuery := bleve.NewMatchQuery(Ngrams(query))
	ngramQuery.SetField("ngrams")
	req := bleve.NewSearchRequestOptions(bleve.NewDisjunctionQuery(textQuery, ngramQuery), limit, 0, false)
	req.Fields = []string{"id"}

	res, err := idx.index.SearchInContext(ctx, req)
	if err != nil {
		return nil, err
	}
	hits := make([]SearchHit, 0, len(res.Hits))
	for _, hit := range res.Hits {
		id, err := parseDocID(hit.ID)
		if err != nil {
			continue
		}
		hits = append(hits, SearchHit{ID: id, Score: hit.Score})
	}
	return hits, nil
}

func Ngrams(text string) string {
	runes := make([]rune, 0, len(text))
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			runes = append(runes, r)
		} else {
			runes = append(runes, ' ')
		}
	}
	words := strings.Fields(string(runes))
	tokens := make([]string, 0)
	for _, word := range words {
		rs := []rune(word)
		if len(rs) <= 1 {
			tokens = append(tokens, word)
			continue
		}
		for i := 0; i < len(rs)-1; i++ {
			tokens = append(tokens, string(rs[i:i+2]))
		}
		tokens = append(tokens, word)
	}
	return strings.Join(tokens, " ")
}

func docID(id int64) string {
	return fmt.Sprintf("article:%d", id)
}

func parseDocID(id string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(id, "article:%d", &n)
	return n, err
}

func stripHTML(input string) string {
	var b strings.Builder
	inTag := false
	for _, r := range input {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}
