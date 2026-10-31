package services

import (
	"encoding/json"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
)

// Microsite template IDs
const (
	MicrositeTemplateBlank     = "blank"
	MicrositeTemplateExecutive = "executive_card"
	MicrositeTemplateCreator   = "creator_links"
	MicrositeTemplateProduct   = "product_showcase"
	MicrositeTemplateEvent     = "event_page"
)

// MicrositeTemplates lists the starter templates offered in the builder.
var MicrositeTemplates = []models.MicrositeTemplateInfo{
	{ID: MicrositeTemplateExecutive, Name: "Digital Business Card", Description: "A letterpress card with painted edges: your name, one-tap Save Contact, then call, WhatsApp and email on the back."},
	{ID: MicrositeTemplateCreator, Name: "Link-in-Bio", Description: "A departure board for your links: the newest one is boarding now."},
	{ID: MicrositeTemplateProduct, Name: "Product Showcase", Description: "A weekly flyer: big prices, deal badges and buy buttons."},
	{ID: MicrositeTemplateEvent, Name: "Event", Description: "A pocket programme: date, venue, countdown, timetable, tickets and add-to-calendar."},
	{ID: MicrositeTemplateBlank, Name: "Blank", Description: "Start from an empty page in the classic style."},
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

func styledTheme(style, font, accent, card, text, buttonText string) models.MicrositeTheme {
	t := DefaultMicrositeTheme()
	t.Style, t.Font = style, font
	t.Accent, t.CardFill, t.Text, t.ButtonText = accent, card, text, buttonText
	t.Background = models.MicrositeBackground{Type: "solid", Color: card}
	return t
}

// BuildMicrositeTemplate returns the pre-filled sample content for a template.
// timezone (IANA) seeds event times; it falls back to UTC.
func BuildMicrositeTemplate(templateID, displayName, timezone string) models.MicrositeContent {
	if displayName == "" {
		displayName = "Your Name"
	}
	if _, err := time.LoadLocation(timezone); err != nil || timezone == "" || timezone == "Local" {
		timezone = "UTC"
	}
	content := models.MicrositeContent{Theme: DefaultMicrositeTheme(), Blocks: []models.MicrositeBlock{}}

	switch templateID {
	case MicrositeTemplateExecutive:
		content.Theme = styledTheme(MicrositeStyleLetterpress, "schibsted", "#2340d6", "#ffffff", "#14161c", "#ffffff")
		content.Blocks = []models.MicrositeBlock{
			mustBlock("blk_profile", models.BlockTypeProfileHeader, models.ProfileHeaderData{
				DisplayName: displayName,
				Title:       "Your Title",
				Company:     "Your Company",
				Bio:         "One line on what you do and how you can help.",
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
		content.Theme = styledTheme(MicrositeStyleBoard, "martian", "#ffd400", "#1b2a6b", "#ffffff", "#1b2a6b")
		content.Blocks = []models.MicrositeBlock{
			mustBlock("blk_profile", models.BlockTypeProfileHeader, models.ProfileHeaderData{
				DisplayName: displayName,
				Bio:         "Creator · Writer · Builder. New stuff every week.",
			}),
			mustBlock("blk_links", models.BlockTypeLinkList, models.LinkListData{Items: []models.LinkItem{
				{ID: "itm_latest", Label: "My latest video", URL: "https://www.youtube.com/", Highlight: true},
				{ID: "itm_newsletter", Label: "Join my newsletter", URL: "https://example.com/newsletter"},
				{ID: "itm_shop", Label: "Shop my favourites", URL: "https://example.com/shop"},
				{ID: "itm_podcast", Label: "The podcast", URL: "https://example.com/podcast"},
			}}),
			mustBlock("blk_social", models.BlockTypeSocialIcons, models.SocialIconsData{Items: []models.SocialLink{
				{ID: "itm_instagram", Platform: "instagram", URL: "https://www.instagram.com/"},
				{ID: "itm_youtube", Platform: "youtube", URL: "https://www.youtube.com/"},
				{ID: "itm_tiktok", Platform: "tiktok", URL: "https://www.tiktok.com/"},
			}}),
		}

	case MicrositeTemplateProduct:
		content.Theme = styledTheme(MicrositeStyleFlyer, "barlow", "#d7261e", "#ffffff", "#17181c", "#ffffff")
		content.Blocks = []models.MicrositeBlock{
			mustBlock("blk_profile", models.BlockTypeProfileHeader, models.ProfileHeaderData{
				DisplayName: displayName,
				Title:       "Open daily 9–6",
				Bio:         "This week's deals",
			}),
			mustBlock("blk_hero", models.BlockTypeProductCard, models.ProductItem{
				Name: "Your best seller", Price: "$24.99", ComparePrice: "$31.00", Badge: "Save 20%",
				Description: "Say what it is and why people love it.", BuyURL: "https://example.com/buy", ButtonText: "Buy now",
			}),
			mustBlock("blk_p2", models.BlockTypeProductCard, models.ProductItem{
				Name: "Second product", Price: "$12.50", BuyURL: "https://example.com/buy-2", ButtonText: "Buy",
			}),
			mustBlock("blk_p3", models.BlockTypeProductCard, models.ProductItem{
				Name: "Third product", Price: "$9.00", Badge: "New", BuyURL: "https://example.com/buy-3", ButtonText: "Buy",
			}),
			mustBlock("blk_week", models.BlockTypeProductCarousel, models.ProductCarouselData{Title: "Also this week", Items: []models.ProductItem{
				{ID: "itm_a", Name: "Bundle deal", Price: "$40.00", BuyURL: "https://example.com/bundle"},
				{ID: "itm_b", Name: "Gift card", Price: "$25.00", BuyURL: "https://example.com/gift"},
				{ID: "itm_c", Name: "Sampler", Price: "$15.00", BuyURL: "https://example.com/sampler"},
			}}),
			mustBlock("blk_contact", models.BlockTypeContactButtons, models.ContactButtonsData{Items: []models.ContactButton{
				{ID: "itm_whatsapp", Kind: "whatsapp", Value: "+10000000000", Label: "Order on WhatsApp"},
				{ID: "itm_call", Kind: "call", Value: "+10000000000", Label: "Call the shop"},
			}}),
		}

	case MicrositeTemplateEvent:
		loc, _ := time.LoadLocation(timezone)
		day := time.Now().In(loc).AddDate(0, 1, 0)
		at := func(h, m int) string {
			return time.Date(day.Year(), day.Month(), day.Day(), h, m, 0, 0, loc).Format(eventTimeLayout)
		}
		// The page name is usually the event's name; the header is the organiser.
		eventTitle := displayName
		if eventTitle == "Your Name" {
			eventTitle = "Your Event Name"
		}
		content.Theme = styledTheme(MicrositeStyleProgramme, "hanken", "#0e7c66", "#ffffff", "#111318", "#ffffff")
		content.Blocks = []models.MicrositeBlock{
			mustBlock("blk_profile", models.BlockTypeProfileHeader, models.ProfileHeaderData{DisplayName: "Your Organisation"}),
			mustBlock("blk_event", models.BlockTypeEventDetails, models.EventDetailsData{
				Title: eventTitle, Start: at(18, 0), End: at(21, 30), Timezone: timezone,
				Venue: "Venue name", Address: "Street, City",
				TicketURL: "https://example.com/tickets", TicketText: "Get tickets", ShowAddToCalendar: true,
			}),
			mustBlock("blk_countdown", models.BlockTypeCountdown, models.CountdownData{
				Target: at(18, 0), Timezone: timezone, Label: "Doors open in", DoneText: "We're live — come on in",
			}),
			mustBlock("blk_agenda", models.BlockTypeAgenda, models.AgendaData{Title: "Programme", Items: []models.AgendaItem{
				{ID: "itm_doors", Time: "18:00", Title: "Doors open", Detail: "Check in at the front desk."},
				{ID: "itm_welcome", Time: "18:30", Title: "Welcome", Speaker: "Host name", Track: "Main stage"},
				{ID: "itm_keynote", Time: "19:00", Title: "Keynote", Speaker: "Speaker name", Track: "Main stage", Detail: "What the talk is about, in two or three lines."},
				{ID: "itm_workshop", Time: "19:45", Title: "Hands-on workshop", Speaker: "Facilitator name", Track: "Room B"},
				{ID: "itm_drinks", Time: "20:30", Title: "Drinks & networking"},
			}}),
			mustBlock("blk_map", models.BlockTypeMap, models.MapData{Label: "Getting there", Address: "Street, City"}),
		}

	default:
		content.Blocks = []models.MicrositeBlock{
			mustBlock("blk_profile", models.BlockTypeProfileHeader, models.ProfileHeaderData{DisplayName: displayName}),
		}
	}
	return content
}
