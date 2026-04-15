//go:build windows

package types

// UIConfig represents UI customization options.
type UIConfig struct {
	DialogStyle    DialogStyle    `json:"DialogStyle"`
	DialogPosition DialogPosition `json:"DialogPosition"`
}

// AssetsConfig represents custom asset paths.
type AssetsConfig struct {
	AppIcon       string `json:"AppIcon"`
	AppIconDark   string `json:"AppIconDark"`
	BannerClassic string `json:"BannerClassic"`
}

// ToolkitConfig represents toolkit-level configuration.
type ToolkitConfig struct {
	UI     UIConfig     `json:"UI"`
	Assets AssetsConfig `json:"Assets"`
}
