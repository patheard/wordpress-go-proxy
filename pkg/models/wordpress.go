package models

import (
	"html/template"
	"log"
	"strings"
)

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

type WordPressMenuItem struct {
	ID    int `json:"id"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Parent int    `json:"parent"`
	Url    string `json:"url"`
}

type PageData struct {
	Lang         string
	LangSwapPath string
	LangSwapSlug string
	Home         string
	Modified     string
	Title        template.HTML
	Content      template.HTML
	Excerpt      template.HTML
	Path         string
	SiteName     string
	Menu         *MenuData
}

type MenuItemData struct {
	ID       int
	Title    string
	Url      string
	Children []*MenuItemData
}

type MenuData struct {
	Items []*MenuItemData
}

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
		Lang:         lang,
		LangSwapPath: langPaths[lang].swap,
		LangSwapSlug: langPaths[lang].slug,
		Home:         langPaths[lang].home,
		Modified:     strings.Split(page.Modified, "T")[0],
		Title:        template.HTML(page.Title.Rendered),
		Content:      template.HTML(strings.ReplaceAll(page.Content.Rendered, baseUrl, "")),
		Excerpt:      template.HTML(page.Excerpt.Rendered),
		SiteName:     siteNames[lang],
		Menu:         menu,
	}
}

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
