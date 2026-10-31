package services

import (
	"encoding/json"

	"github.com/URLshorter/url-shortener/internal/models"
)

// Microsite template IDs
const (
	MicrositeTemplateBlank     = "blank"
	MicrositeTemplateExecutive = "executive_card"
	MicrositeTemplateCreator   = "creator_links"
)

// MicrositeTemplates lists the starter templates offered in the builder.
var MicrositeTemplates = []models.MicrositeTemplateInfo{
	{ID: MicrositeTemplateExecutive, Name: "Digital Business Card", Description: "Avatar, title, company, one-tap Save Contact, call / WhatsApp / email buttons and socials."},
	{ID: MicrositeTemplateCreator, Name: "Link-in-Bio", Description: "Profile, a stack of highlighted links and a social icon row."},
	{ID: MicrositeTemplateBlank, Name: "Blank", Description: "Start from an empty page."},
}

// IsMicrositeTemplate reports whether id is a known template.
func IsMicrositeTemplate(id string) bool {
	for _, t := range MicrositeTemplates {
		if t.ID == id {
			return true
		}
	}
	return false
}

func mustBlock(id, blockType string, data interface{}) models.MicrositeBlock {
	raw, err := json.Marshal(data)
	if err != nil {
		panic(err) // static template data; cannot fail
	}
	return models.MicrositeBlock{ID: id, Type: blockType, SchemaVersion: currentBlockSchema, Data: raw}
}

// BuildMicrositeTemplate returns the pre-filled content for a template.
func BuildMicrositeTemplate(templateID, displayName string) models.MicrositeContent {
	if displayName == "" {
		displayName = "Your Name"
	}
	content := models.MicrositeContent{Theme: DefaultMicrositeTheme(), Blocks: []models.MicrositeBlock{}}

	switch templateID {
	case MicrositeTemplateExecutive:
		content.Blocks = []models.MicrositeBlock{
			mustBlock("blk_profile", models.BlockTypeProfileHeader, models.ProfileHeaderData{
				DisplayName: displayName,
				Title:       "Your Title",
				Company:     "Your Company",
				Bio:         "A short line about what you do and how you can help.",
			}),
			mustBlock("blk_vcard", models.BlockTypeVCardAction, models.VCardActionData{
				ButtonText:   "Save Contact",
				FirstName:    displayName,
				Organization: "Your Company",
				Title:        "Your Title",
			}),
			mustBlock("blk_contact", models.BlockTypeContactButtons, models.ContactButtonsData{Items: []models.ContactButton{
				{ID: "itm_call", Kind: "call", Value: "+10000000000", Label: "Call"},
				{ID: "itm_whatsapp", Kind: "whatsapp", Value: "+10000000000", Label: "WhatsApp"},
				{ID: "itm_email", Kind: "email", Value: "you@example.com", Label: "Email"},
			}}),
			mustBlock("blk_social", models.BlockTypeSocialIcons, models.SocialIconsData{Items: []models.SocialLink{
				{ID: "itm_linkedin", Platform: "linkedin", URL: "https://www.linkedin.com/"},
				{ID: "itm_website", Platform: "website", URL: "https://example.com/"},
			}}),
		}
	case MicrositeTemplateCreator:
		content.Theme.Background = models.MicrositeBackground{Type: "linear", Stops: []string{"#f857a6", "#ff5858"}, Angle: 135}
		content.Theme.Accent = "#111827"
		content.Theme.ButtonRadius = "pill"
		content.Blocks = []models.MicrositeBlock{
			mustBlock("blk_profile", models.BlockTypeProfileHeader, models.ProfileHeaderData{
				DisplayName: displayName,
				Bio:         "Creator · Writer · Builder",
			}),
			mustBlock("blk_links", models.BlockTypeLinkList, models.LinkListData{Items: []models.LinkItem{
				{ID: "itm_latest", Label: "My latest video", URL: "https://www.youtube.com/", Highlight: true},
				{ID: "itm_newsletter", Label: "Join my newsletter", URL: "https://example.com/newsletter"},
				{ID: "itm_shop", Label: "Shop my favourites", URL: "https://example.com/shop"},
			}}),
			mustBlock("blk_social", models.BlockTypeSocialIcons, models.SocialIconsData{Items: []models.SocialLink{
				{ID: "itm_instagram", Platform: "instagram", URL: "https://www.instagram.com/"},
				{ID: "itm_youtube", Platform: "youtube", URL: "https://www.youtube.com/"},
				{ID: "itm_tiktok", Platform: "tiktok", URL: "https://www.tiktok.com/"},
			}}),
		}
	default:
		content.Blocks = []models.MicrositeBlock{
			mustBlock("blk_profile", models.BlockTypeProfileHeader, models.ProfileHeaderData{DisplayName: displayName}),
		}
	}
	return content
}
