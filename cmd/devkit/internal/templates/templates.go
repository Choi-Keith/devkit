// Package templates is the registry of project templates known to devkit.
//
// Templates are NOT embedded in the binary. They live at the repository root
// and are downloaded at generation time (see internal/fetch), so they are
// always up to date and this module never needs a sync step.
package templates

import "fmt"

// DefaultRepo is the git repository templates are downloaded from.
// Template.Name maps to the subdirectory Subdir inside this repo.
const DefaultRepo = "https://github.com/Choi-Keith/devkit"

// Template describes one project template.
type Template struct {
	// Name is the identifier used on the command line, e.g. `devkit new golang-ddd`.
	Name string
	// Subdir is the template's directory inside the repository.
	Subdir string
	// Lang is the primary language of the template.
	Lang string
	// Desc is a short human-readable description.
	Desc string
}

// All is the registry of available templates, in display order.
var All = []Template{
	{
		Name:   "golang-ddd",
		Subdir: "golang-ddd-template",
		Lang:   "Go",
		Desc:   "Go DDD 分层架构模板(configs / deployments / internal 分层)",
	},
	{
		Name:   "golang-feature-mvc",
		Subdir: "golang-feature-mvc-template",
		Lang:   "Go",
		Desc:   "Go Feature MVC 模板(按 feature 纵向切分 + GORM + Redis + JWT + Swagger)",
	},
	{
		Name:   "golang-simple",
		Subdir: "golang-simple-template",
		Lang:   "Go",
		Desc:   "Go 轻量级单体服务模板",
	},
	{
		Name:   "nestjs-simple",
		Subdir: "nestjs-simple-template",
		Lang:   "TypeScript",
		Desc:   "NestJS 11 模板(TypeORM + Redis + JWT + RBAC)",
	},
	{
		Name:   "python-fastapi",
		Subdir: "python-fastapi-template",
		Lang:   "Python",
		Desc:   "FastAPI 模板(Alembic 迁移 + uv 依赖管理)",
	},
}

// Find returns the template with the given name.
func Find(name string) (Template, error) {
	for _, t := range All {
		if t.Name == name {
			return t, nil
		}
	}
	return Template{}, fmt.Errorf("unknown template %q (run `devkit list` to see available templates)", name)
}
