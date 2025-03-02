package models

import (
	"html/template"
	"log"
	"strings"
)

// WordPressPage represents a WordPress page JSON response.
type WordPressPage struct {
	ID       int    `json:"id"`
	Slug     string `json:"slug"`
	SlugEn   string `json:"slug_en"`
	SlugFr   string `json:"slug_fr"`
	Lang     string `json:"lang"`
	Modified string `json:"modified"`
	Content  struct {
		Rendered string `json:"rendered"`
		Raw      string `json:"raw,omitempty"`
	} `json:"content"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Excerpt struct {
		Rendered string `json:"rendered,omitempty"`
	} `json:"excerpt,omitempty"`
	FeaturedMedia int   `json:"featured_media,omitempty"`
	Categories    []int `json:"categories,omitempty"`
}

// WordPressMenuItem represents a WordPress menu item JSON response.
type WordPressMenuItem struct {
	ID    int `json:"id"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Parent int    `json:"parent"`
	Url    string `json:"url"`
}

// PageData holds the data needed to render a page.
type PageData struct {
	Lang           string
	LangSwapPath   string
	LangSwapSlug   string
	Home           string
	Modified       string
	Title          template.HTML
	Content        template.HTML
	ShowBreadcrumb bool
	SiteName       string
	Menu           *MenuData
}

// MenuItemData holds the data needed to render a menu item.
type MenuItemData struct {
	ID       int
	Title    string
	Url      string
	Children []*MenuItemData
}

// MenuData holds the data needed to render a menu.
type MenuData struct {
	Items []*MenuItemData
}

// NewPageData creates a new PageData object that can then be used to render a page.
func NewPageData(page *WordPressPage, menu *MenuData, siteNames map[string]string, baseUrl string) PageData {
	lang := page.Lang
	if lang != "en" && lang != "fr" {
		lang = "en"
		log.Printf("Warning: Invalid language '%s', defaulting to 'en'", page.Lang)
	}

	langPaths := map[string]struct {
		swap string
		slug string
		home string
	}{
		"en": {"/fr/", page.SlugFr, "/"},
		"fr": {"/", page.SlugEn, "/fr/"},
	}

	return PageData{
		Lang:           lang,
		LangSwapPath:   langPaths[lang].swap,
		LangSwapSlug:   langPaths[lang].slug,
		Home:           langPaths[lang].home,
		Modified:       strings.Split(page.Modified, "T")[0],
		Title:          template.HTML(page.Title.Rendered),
		Content:        template.HTML(strings.ReplaceAll(page.Content.Rendered, baseUrl, "")),
		ShowBreadcrumb: !strings.Contains(page.Slug, "home"),
		SiteName:       siteNames[lang],
		Menu:           menu,
	}
}

// NewMenuData creates a new MenuData object that can then be used to render a menu.
// The menu items are expected to be in a flat list with parent/child relationships
// represented by the Parent field.
func NewMenuData(menuItems *[]WordPressMenuItem, baseUrl string) *MenuData {
	menuMap := make(map[int]*MenuItemData)
	for _, item := range *menuItems {
		menuMap[item.ID] = &MenuItemData{
			ID:       item.ID,
			Title:    item.Title.Rendered,
			Url:      strings.Replace(item.Url, baseUrl, "", 1),
			Children: make([]*MenuItemData, 0),
		}
	}

	// Build up the menu tree of parent/child relationships
	menuTree := make([]*MenuItemData, 0)
	for _, item := range *menuItems {
		if item.Parent != 0 {
			if parent, ok := menuMap[item.Parent]; ok {
				parent.Children = append(parent.Children, menuMap[item.ID])
			}
		} else {
			menuTree = append(menuTree, menuMap[item.ID])
		}
	}

	return &MenuData{
		Items: menuTree,
	}
}
