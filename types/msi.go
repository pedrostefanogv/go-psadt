//go:build windows

package types

// MsiTableOptions options for Get-ADTMsiTableProperty.
type MsiTableOptions struct {
	Path                        string `ps:"Path"`
	Table                       string `ps:"Table"`
	TablePropertyNameColumnNum  int    `ps:"TablePropertyNameColumnNum"`
	TablePropertyValueColumnNum int    `ps:"TablePropertyValueColumnNum"`
}

// SetMsiPropertyOptions options for Set-ADTMsiProperty.
type SetMsiPropertyOptions struct {
	DataBase      interface{} `ps:"DataBase"`
	PropertyName  string      `ps:"PropertyName"`
	PropertyValue string      `ps:"PropertyValue"`
}

// MsiTransformOptions options for New-ADTMsiTransform.
type MsiTransformOptions struct {
	MsiPath    string            `ps:"MsiPath"`
	MstPath    string            `ps:"MstPath"`
	Transforms map[string]string `ps:"Transforms"`
}
