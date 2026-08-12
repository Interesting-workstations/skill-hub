package skill

import (
	"net/http"
	"strconv"

	"github.com/Interesting-workstations/skill-hub/server/internal/response"
)

// Handler 负责 HTTP 层：参数解析、调用 Service、返回统一响应。
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 注册技能资源库相关路由。
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/health", h.health)
	mux.HandleFunc("GET /api/v1/stats", h.stats)
	mux.HandleFunc("GET /api/v1/skills", h.listSkills)
	mux.HandleFunc("GET /api/v1/skills/{id}", h.getSkill)
	mux.HandleFunc("GET /api/v1/authors", h.listAuthors)
	mux.HandleFunc("GET /api/v1/authors/{slug}", h.getAuthor)
	mux.HandleFunc("GET /api/v1/categories", h.listCategories)
	mux.HandleFunc("GET /api/v1/categories/{slug}", h.getCategory)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, map[string]string{"status": "ok"})
}

// GET /api/v1/stats
func (h *Handler) stats(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, h.svc.Stats())
}

// GET /api/v1/skills?category=&author=&official=&featured=
func (h *Handler) listSkills(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	official, err1 := strconv.ParseBool(q.Get("official"))
	featured, err2 := strconv.ParseBool(q.Get("featured"))
	filter := SkillFilter{
		Category: q.Get("category"),
		Author:   q.Get("author"),
	}
	if err1 == nil {
		filter.Official = official
	}
	if err2 == nil {
		filter.Featured = featured
	}
	response.OK(w, h.svc.ListSkills(filter))
}

// GET /api/v1/skills/{id}
func (h *Handler) getSkill(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	skill, ok := h.svc.GetSkill(id)
	if !ok {
		response.NotFound(w, "技能不存在")
		return
	}
	response.OK(w, skill)
}

// GET /api/v1/authors
func (h *Handler) listAuthors(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, h.svc.ListAuthors())
}

// GET /api/v1/authors/{slug}
func (h *Handler) getAuthor(w http.ResponseWriter, r *http.Request) {
	detail, ok := h.svc.GetAuthor(r.PathValue("slug"))
	if !ok {
		response.NotFound(w, "作者不存在")
		return
	}
	response.OK(w, detail)
}

// GET /api/v1/categories
func (h *Handler) listCategories(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, h.svc.ListCategories())
}

// GET /api/v1/categories/{slug}
func (h *Handler) getCategory(w http.ResponseWriter, r *http.Request) {
	cat, ok := h.svc.GetCategory(r.PathValue("slug"))
	if !ok {
		response.NotFound(w, "分类不存在")
		return
	}
	response.OK(w, cat)
}
