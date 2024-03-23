package main

import (
	"context"
	"fmt"
	"log"
	"notion-watchlist-integration/cmd/mediaInfoProviders"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/jomei/notionapi"
)

const (
	notionAPIIdentifier        = "NOTION_API_SECRET" // sometimes called "Token"
	notionDatabaseIDIdentifier = "NOTION_DATABASE"
)

func PropMediaType(props notionapi.Properties) mediaInfoProviders.MediaInfoProvider {
	mediaType, ok := props["Media Type"].(*notionapi.SelectProperty)
	if !ok {
		return nil
	}
	var mediaInfoProvider mediaInfoProviders.MediaInfoProvider
	switch mediaType.Select.Name {
	case "Book":
		mediaInfoProvider = &mediaInfoProviders.GoogleBooksMediaInfo{}
	case "Film":
		// TODO: IMDB
	case "Game":
		// TODO: steam??
	}
	return mediaInfoProvider
}

func PopulatePropsWithMediaInfo(ctx context.Context, info *mediaInfoProviders.MediaInfo) (notionapi.Properties, error) {
	props := notionapi.Properties{}
	log.Printf("%+v\n", props)
	if len(info.Authors) > 0 {
		props["Author/Director"] = &notionapi.TextProperty{
			Text: []notionapi.RichText{
				{
					PlainText: strings.Join(info.Authors, ","),
				},
			},
		}
	}
	if info.Summary != "" {
		props["Summary"] = &notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Text: &notionapi.Text{
						Content: info.Summary,
					},
				},
			},
		}
	}
	if len(info.Category) > 0 {
		thing := &notionapi.MultiSelectProperty{
			MultiSelect: []notionapi.Option{},
		}
		for _, v := range info.Category {
			thing.MultiSelect = append(thing.MultiSelect, notionapi.Option{Name: v})
		}
		props["Category"] = thing
	}
	if info.Rating > 0 {
		s := "⭐️"
		props["Avg. Rating"] = &notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: strings.Repeat(s, int(info.Rating)),
			},
		}
	}
	if info.PageCount > 0 {
		props["Total Pages"] = &notionapi.NumberProperty{
			Number: float64(info.PageCount),
		}
	}
	log.Printf("%+v\n", props)
	return props, nil
}

func main() {
	err := godotenv.Load()
	if err != nil && false {
		log.Fatal("Error loading .env file")
	}
	ctx := context.Background()

	notionAPIKey := notionapi.Token(os.Getenv(notionAPIIdentifier))
	client := notionapi.NewClient(notionAPIKey)

	log.SetFlags(log.Lshortfile)

	dbid := notionapi.DatabaseID(os.Getenv(notionDatabaseIDIdentifier))
	log.Println(client)

	res, err := client.Database.Query(ctx, dbid, &notionapi.DatabaseQueryRequest{
		PageSize: 100,
	})
	if err != nil {
		log.Fatalf("Unable to get db %v: %v", dbid, err)
	}
	for _, page := range res.Results {
		if page.Object != notionapi.ObjectTypePage {
			continue // not page-- confusing
		}
		t := page.Properties["Title"].(*notionapi.TitleProperty)
		if len(t.Title) != 1 {
			log.Printf("bad number of titles: %d", len(t.Title))
			continue
		}
		title := t.Title[0].Text.Content
		// log.Printf("%+v\n", title)
		v, ok := page.Properties["noauto"].(*notionapi.CheckboxProperty)
		if !ok {
			log.Println("`noauto` field not found")
		}
		if v.Checkbox {
			continue
		}

		mediaInfoProvider := PropMediaType(page.Properties)
		if mediaInfoProvider == nil {
			log.Println(fmt.Errorf("no media provider found"))
			continue
		}
		info, err := mediaInfoProvider.GetMediaInfo(ctx, title)
		if err != nil {
			log.Println(err)
			continue
		}
		props, err := PopulatePropsWithMediaInfo(ctx, info)
		if err != nil {
			log.Println(err)
			continue
		}

		pageID := notionapi.PageID(page.ID)
		log.Printf("%+v", props)
		page, err := client.Page.Update(ctx, pageID, &notionapi.PageUpdateRequest{
			Properties: props,
			Cover: &notionapi.Image{
				External: &notionapi.FileObject{URL: info.Image},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Println(props, page)
	}
}
