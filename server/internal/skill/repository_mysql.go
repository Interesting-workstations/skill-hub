package skill

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
	"github.com/go-sql-driver/mysql"
)

// mysqlRepo 基于 MySQL 数据库的实现。
// 首次启动时若库中无数据，从种子 JSON 初始化。
type mysqlRepo struct {
	db *sql.DB
}

// NewMySQLRepository 连接 MySQL 数据库并确保表结构与种子数据就绪。
// DSN 示例：user:pass@tcp(127.0.0.1:3306)/skillhub?charset=utf8mb4&parseTime=true
// 目标数据库不存在时自动创建（utf8mb4）。
func NewMySQLRepository(dsn, seedJSONPath string) (Repository, error) {
	if err := ensureDatabase(dsn); err != nil {
		return nil, err
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	repo := &mysqlRepo{db: db}
	if err := repo.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	if seedJSONPath != "" {
		if err := repo.seedIfEmpty(seedJSONPath); err != nil {
			db.Close()
			return nil, err
		}
	}
	return repo, nil
}

// ReplaceAll 清空数据库全部表并用 store 重建（用于导入爬取数据）。
func ReplaceAll(dsn string, store domain.Store) error {
	repo, err := NewMySQLRepository(dsn, "")
	if err != nil {
		return err
	}
	return repo.(*mysqlRepo).replaceAll(store)
}

// ensureDatabase 解析 DSN，目标数据库不存在时自动创建（utf8mb4）。
func ensureDatabase(dsn string) error {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("解析 DSN 失败: %w", err)
	}
	if cfg.DBName == "" {
		return fmt.Errorf("DSN 必须指定数据库名（如 /skillhub）")
	}
	serverCfg := *cfg
	serverCfg.DBName = ""
	server, err := sql.Open("mysql", serverCfg.FormatDSN())
	if err != nil {
		return err
	}
	defer server.Close()
	_, err = server.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.DBName,
	))
	return err
}

// migrate 创建数据表（逐条执行，MySQL 驱动默认不支持多语句）。
func (r *mysqlRepo) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS categories (
			slug VARCHAR(100) PRIMARY KEY,
			name VARCHAR(255) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS authors (
			slug VARCHAR(100) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			avatar VARCHAR(64) NOT NULL,
			skill_count INT NOT NULL DEFAULT 0
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS skills (
			id VARCHAR(150) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			author VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,
			category VARCHAR(100) NOT NULL,
			download_url VARCHAR(500) NOT NULL DEFAULT '',
			is_official TINYINT(1) NOT NULL DEFAULT 0,
			is_featured TINYINT(1) NOT NULL DEFAULT 0,
			install_command TEXT,
			github_url VARCHAR(500),
			github_stars VARCHAR(64),
			license VARCHAR(128),
			skill_path VARCHAR(500) NOT NULL DEFAULT '',
			tags TEXT NOT NULL,
			content MEDIUMTEXT NOT NULL,
			data_status VARCHAR(20) NOT NULL DEFAULT 'published',
			KEY idx_skills_category (category),
			KEY idx_skills_author (author)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS article_views (
			article_id VARCHAR(150) NOT NULL,
			ip VARCHAR(64) NOT NULL,
			viewed_date DATE NOT NULL,
			PRIMARY KEY (article_id, ip, viewed_date)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, stmt := range stmts {
		if _, err := r.db.Exec(stmt); err != nil {
			return err
		}
	}
	if err := r.ensureDataStatusColumn(); err != nil {
		return err
	}
	return r.ensureSkillPathColumn()
}

// ensureSkillPathColumn 给已存在的 skills 表补充 skill_path（技能目录）列。
func (r *mysqlRepo) ensureSkillPathColumn() error {
	if _, err := r.db.Exec(`ALTER TABLE skills ADD COLUMN skill_path VARCHAR(500) NOT NULL DEFAULT ''`); err != nil {
		if strings.Contains(err.Error(), "Duplicate column") {
			return nil
		}
		return err
	}
	return nil
}

// ensureDataStatusColumn 给已存在的 skills 表补充 data_status（审核状态）列。
func (r *mysqlRepo) ensureDataStatusColumn() error {
	if _, err := r.db.Exec(`ALTER TABLE skills ADD COLUMN data_status VARCHAR(20) NOT NULL DEFAULT 'published'`); err != nil {
		if strings.Contains(err.Error(), "Duplicate column") {
			return nil
		}
		return err
	}
	return nil
}

// seedIfEmpty 当 skills 表为空时，从种子 JSON 初始化数据。
func (r *mysqlRepo) seedIfEmpty(seedJSONPath string) error {
	var n int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM skills`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	data, err := os.ReadFile(seedJSONPath)
	if err != nil {
		return fmt.Errorf("读取种子数据失败: %w", err)
	}
	var store domain.Store
	if err := json.Unmarshal(data, &store); err != nil {
		return fmt.Errorf("解析种子数据失败: %w", err)
	}
	return r.insertStore(store)
}

// insertStore 在单个事务中写入作者、分类与技能。
func (r *mysqlRepo) insertStore(store domain.Store) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := r.insertStoreTx(tx, store); err != nil {
		return err
	}
	return tx.Commit()
}

// replaceAll 清空全部数据表并用 store 重建。
func (r *mysqlRepo) replaceAll(store domain.Store) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, table := range []string{"skills", "authors", "categories"} {
		if _, err := tx.Exec(`DELETE FROM ` + table); err != nil {
			return err
		}
	}
	if err := r.insertStoreTx(tx, store); err != nil {
		return err
	}
	return tx.Commit()
}

// insertStoreTx 向事务写入作者、分类与技能。
func (r *mysqlRepo) insertStoreTx(tx *sql.Tx, store domain.Store) error {
	for _, a := range store.Authors {
		if _, err := tx.Exec(
			`INSERT IGNORE INTO authors(slug, name, avatar, skill_count) VALUES(?,?,?,?)`,
			a.Slug, a.Name, a.Avatar, a.SkillCount,
		); err != nil {
			return err
		}
	}
	for _, c := range store.SkillCategories {
		if _, err := tx.Exec(
			`INSERT IGNORE INTO categories(slug, name) VALUES(?,?)`,
			c.Slug, c.Name,
		); err != nil {
			return err
		}
	}
	all := append(append([]domain.Skill{}, store.FeaturedSkills...), flattenCategorySkills(store.SkillCategories)...)
	// 批内去重：ID 或同源同名（name+author+githubUrl）相同只保留第一条，防止重复入库
	seenID := make(map[string]bool)
	seenSource := make(map[string]bool)
	for _, s := range all {
		if seenID[s.ID] {
			continue
		}
		sourceKey := s.Name + "|" + s.Author + "|" + s.GithubURL
		if seenSource[sourceKey] {
			continue
		}
		seenID[s.ID] = true
		seenSource[sourceKey] = true
		if err := insertSkill(tx, s); err != nil {
			return err
		}
	}
	return nil
}

// insertSkill 插入一条技能记录。
func insertSkill(tx *sql.Tx, s domain.Skill) error {
	tags, err := json.Marshal(s.Tags)
	if err != nil {
		return err
	}
	content, err := json.Marshal(s.Content)
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		`INSERT IGNORE INTO skills(
			id, name, author, description, category, download_url,
			is_official, is_featured, install_command, github_url,
			github_stars, license, skill_path, tags, content
		) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		s.ID, s.Name, s.Author, s.Description, s.Category, s.DownloadURL,
		boolInt(s.IsOfficial), boolInt(s.IsFeatured), s.InstallCommand, s.GithubURL,
		s.GithubStars, s.License, s.SkillPath, string(tags), string(content),
	)
	return err
}

// AllSkills 返回全部技能。
func (r *mysqlRepo) AllSkills() []domain.Skill {
	rows, err := r.db.Query(`SELECT id, name, author, description, category,
		download_url, is_official, is_featured, install_command, github_url,
		github_stars, license, skill_path, tags, content FROM skills`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	skills := make([]domain.Skill, 0, 64)
	for rows.Next() {
		if s, ok := scanSkill(rows); ok {
			skills = append(skills, s)
		}
	}
	return skills
}

// SkillByID 按 ID 查询技能。
func (r *mysqlRepo) SkillByID(id string) (domain.Skill, bool) {
	row := r.db.QueryRow(`SELECT id, name, author, description, category,
		download_url, is_official, is_featured, install_command, github_url,
		github_stars, license, skill_path, tags, content FROM skills WHERE id = ?`, id)
	s, ok := scanSkill(row)
	return s, ok
}

// AllAuthors 返回全部作者；SkillCount 为该作者下实际技能数、
// OfficialSkills 为官方技能数（均实时统计）。
func (r *mysqlRepo) AllAuthors() []domain.Author {
	rows, err := r.db.Query(`
		SELECT a.slug, a.name, a.avatar,
			COUNT(s.id),
			COALESCE(SUM(s.is_official), 0)
		FROM authors a
		LEFT JOIN skills s ON s.author = a.slug
		GROUP BY a.slug, a.name, a.avatar
		ORDER BY a.slug`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	authors := make([]domain.Author, 0, 16)
	for rows.Next() {
		var a domain.Author
		if err := rows.Scan(&a.Slug, &a.Name, &a.Avatar, &a.SkillCount, &a.OfficialSkills); err != nil {
			continue
		}
		authors = append(authors, a)
	}
	return authors
}

// AllCategories 返回全部分类；count 为该分类下实际技能数（实时统计）。
func (r *mysqlRepo) AllCategories() []domain.Category {
	rows, err := r.db.Query(`
		SELECT c.slug, c.name, COUNT(s.id)
		FROM categories c
		LEFT JOIN skills s ON s.category = c.slug
		GROUP BY c.slug, c.name
		ORDER BY c.slug`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.Slug, &c.Name, &c.Count); err != nil {
			continue
		}
		c.Skills = r.skillsByCategory(c.Slug)
		categories = append(categories, c)
	}
	return categories
}

func (r *mysqlRepo) skillsByCategory(slug string) []domain.Skill {
	rows, err := r.db.Query(`SELECT id, name, author, description, category,
		download_url, is_official, is_featured, install_command, github_url,
		github_stars, license, skill_path, tags, content FROM skills WHERE category = ?`, slug)
	if err != nil {
		return nil
	}
	defer rows.Close()

	skills := make([]domain.Skill, 0, 8)
	for rows.Next() {
		if s, ok := scanSkill(rows); ok {
			skills = append(skills, s)
		}
	}
	return skills
}

// scanSkill 将查询行扫描为 domain.Skill（tags/content 为 JSON 文本列）。
type rowScanner interface {
	Scan(dest ...any) error
}

func scanSkill(scanner rowScanner) (domain.Skill, bool) {
	var s domain.Skill
	var official, featured int
	var tagsRaw, contentRaw string
	err := scanner.Scan(
		&s.ID, &s.Name, &s.Author, &s.Description, &s.Category,
		&s.DownloadURL, &official, &featured, &s.InstallCommand, &s.GithubURL,
		&s.GithubStars, &s.License, &s.SkillPath, &tagsRaw, &contentRaw,
	)
	if err != nil {
		return domain.Skill{}, false
	}
	s.IsOfficial = official == 1
	s.IsFeatured = featured == 1
	_ = json.Unmarshal([]byte(tagsRaw), &s.Tags)
	_ = json.Unmarshal([]byte(contentRaw), &s.Content)
	return s, true
}

func flattenCategorySkills(categories []domain.Category) []domain.Skill {
	var skills []domain.Skill
	for _, c := range categories {
		skills = append(skills, c.Skills...)
	}
	return skills
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ---------- 公开内容（文章 / 站点配置 / SEO / 提交技能） ----------

// ListArticles 返回全部已发布文章（按更新时间倒序）。
func (r *mysqlRepo) ListArticles() []domain.Article {
	rows, err := r.db.Query(`SELECT id, title, status, category, author, views, updated_at, content
		FROM articles WHERE status = 'published' ORDER BY updated_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]domain.Article, 0, 8)
	for rows.Next() {
		var a domain.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Status, &a.Category, &a.Author, &a.Views, &a.UpdatedAt, &a.Content); err != nil {
			continue
		}
		out = append(out, a)
	}
	return out
}

// ArticleByID 按 ID 查询已发布文章；浏览量去重 +1：
//   - countView=false：不计数（前端已读过的文章，配合 ?incr=0 跳过）。
//   - 同一 IP 当天对同一文章只计一次（article_views 表 + INSERT IGNORE 原子判断）。
func (r *mysqlRepo) ArticleByID(id, ip string, countView bool) (domain.Article, bool) {
	var a domain.Article
	err := r.db.QueryRow(`SELECT id, title, status, category, author, views, updated_at, content
		FROM articles WHERE id = ? AND status = 'published'`, id).
		Scan(&a.ID, &a.Title, &a.Status, &a.Category, &a.Author, &a.Views, &a.UpdatedAt, &a.Content)
	if err != nil {
		return domain.Article{}, false
	}
	if countView && ip != "" {
		res, err := r.db.Exec(
			`INSERT IGNORE INTO article_views (article_id, ip, viewed_date) VALUES (?, ?, CURDATE())`,
			id, ip)
		if err == nil {
			// 仅当首次插入成功（当天该 IP 未访问过）才累加浏览量
			if n, _ := res.RowsAffected(); n > 0 {
				_, _ = r.db.Exec(`UPDATE articles SET views = views + 1 WHERE id = ?`, id)
			}
		}
	}
	return a, true
}

// GetSiteConfig 返回站点配置（无记录时返回默认值）。
func (r *mysqlRepo) GetSiteConfig() (domain.SiteConfig, bool) {
	var s domain.SiteConfig
	err := r.db.QueryRow(`SELECT site_name, slogan, portal_url, icp, contact_email FROM site_config WHERE id = 1`).
		Scan(&s.SiteName, &s.Slogan, &s.PortalUrl, &s.ICP, &s.ContactEmail)
	if err != nil {
		return domain.SiteConfig{}, false
	}
	return s, true
}

// GetSeo 返回 SEO 配置（无记录时返回默认值）。
func (r *mysqlRepo) GetSeo() (domain.SeoConfig, bool) {
	var s domain.SeoConfig
	err := r.db.QueryRow(`SELECT title, description, keywords, og_image FROM seo_config WHERE id = 1`).
		Scan(&s.Title, &s.Description, &s.Keywords, &s.OgImage)
	if err != nil {
		return domain.SeoConfig{}, false
	}
	return s, true
}

// SubmitSkill 保存用户提交的技能：插入 skills 表并标记为待审核（data_status='pending'）。
func (r *mysqlRepo) SubmitSkill(s *domain.Skill) error {
	tags, err := json.Marshal(s.Tags)
	if err != nil {
		return err
	}
	content, err := json.Marshal(s.Content)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(
		`INSERT INTO skills(
			id, name, author, description, category, download_url,
			is_official, is_featured, install_command, github_url,
			github_stars, license, skill_path, tags, content, data_status
		) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		s.ID, s.Name, s.Author, s.Description, s.Category, s.DownloadURL,
		boolInt(s.IsOfficial), boolInt(s.IsFeatured), s.InstallCommand, s.GithubURL,
		s.GithubStars, s.License, s.SkillPath, string(tags), string(content), "pending",
	)
	return err
}
