package other_link

const otherLinkKey = "__thanks_with_ui_other_link__"

type OtherLinks struct {
	Tops     []string            `json:"tops"`
	Groups   []string            `json:"groups"`
	GroupMap map[string][]string `json:"group_map"`
	Links    []OtherLink         `json:"links"`
}

type OtherLink struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Url     string `json:"url"`
	Icon    string `json:"icon"`
	IconUrl string `json:"icon_url"`
}
