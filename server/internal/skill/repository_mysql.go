package skill

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

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
	if err := repo.seedIfEmpty(seedJSONPath); err != nil {
		db.Close()
		return nil, err
	}
	return repo, nil
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
			tags TEXT NOT NULL,
			content MEDIUMTEXT NOT NULL,
			KEY idx_skills_category (category),
			KEY idx_skills_author (author)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, stmt := range stmts {
		if _, err := r.db.Exec(stmt); err != nil {
			return err
		}
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

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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
	for _, s := range all {
		if err := insertSkill(tx, s); err != nil {
			return err
		}
	}
	return tx.Commit()
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
			github_stars, license, tags, content
		) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		s.ID, s.Name, s.Author, s.Description, s.Category, s.DownloadURL,
		boolInt(s.IsOfficial), boolInt(s.IsFeatured), s.InstallCommand, s.GithubURL,
		s.GithubStars, s.License, string(tags), string(content),
	)
	return err
}

// AllSkills 返回全部技能。
func (r *mysqlRepo) AllSkills() []domain.Skill {
	rows, err := r.db.Query(`SELECT id, name, author, description, category,
		download_url, is_official, is_featured, install_command, github_url,
		github_stars, license, tags, content FROM skills`)
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
		github_stars, license, tags, content FROM skills WHERE id = ?`, id)
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
		github_stars, license, tags, content FROM skills WHERE category = ?`, slug)
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
		&s.GithubStars, &s.License, &tagsRaw, &contentRaw,
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
