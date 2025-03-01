package models

import (
	"html/template"
	"strings"
)

type WordPressPage struct {
	ID       int    `json:"id"`
	Slug     string `json:"slug"`
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
	Lang     string
	Modified string
	Title    template.HTML
	Content  template.HTML
	Excerpt  template.HTML
	Path     string
	Menu     *MenuData
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

func NewPageData(page *WordPressPage, menu *MenuData) PageData {
	return PageData{
		Lang:     page.Lang,
		Modified: strings.Split(page.Modified, "T")[0],
		Title:    template.HTML(page.Title.Rendered),
		Content:  template.HTML(page.Content.Rendered),
		Excerpt:  template.HTML(page.Excerpt.Rendered),
		Menu:     menu,
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
