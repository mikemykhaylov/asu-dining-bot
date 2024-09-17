package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/mikemykhaylov/asu-dining-bot/internal/api"
	"github.com/mikemykhaylov/asu-dining-bot/internal/config"
	"github.com/mikemykhaylov/asu-dining-bot/internal/logger"
	"github.com/spf13/viper"
)

var (
	asuDiningWebsiteURL  string = "https://asu.campusdish.com/DiningVenues/Tempe-Campus/Barrett-Dining-Center"
	noMealsAvailableText string = "There are currently no menus available for this meal period and date."
	periodName           string = "Dinner"

	menuWrapper         string = ".MenuWrapperDaily"
	mealSelectionButton string = "button.DateMealFilterButton"
	mealInput           string = "#aria-meal-input"
	doneButton          string = ".Done"
)

type Dishes struct {
	HomeZone    []string
	TrueBalance []string
}

func (d *Dishes) ToString() string {
	out := "<b>Home Zone:</b>\n"
	for _, dish := range d.HomeZone {
		out += fmt.Sprintf("— %s\n", dish)
	}
	out += "\n<b>True Balance:</b>\n"
	for _, dish := range d.TrueBalance {
		out += fmt.Sprintf("— %s\n", dish)
	}
	return out
}

type RunHandler struct {
	TelegramAPI *api.TelegramAPI
}

func NewRunHandler() *RunHandler {
	httpClient := &http.Client{}

	telegramAPI := api.NewTelegramAPI(viper.GetString(config.TelegramBotTokenKey), httpClient)

	return &RunHandler{
		TelegramAPI: telegramAPI,
	}
}

func (r *RunHandler) Run(ctx context.Context) error {
	log := logger.FromContext(ctx)
	log.Info("Running handler")
	defer log.Info("Finished running handler")

	personalID := viper.GetInt64(config.PersonalIDKey)

	var page *rod.Page
	if viper.GetBool(config.DockerKey) {
		u := launcher.New().Bin("/usr/bin/chromium-browser").Headless(true).MustLaunch()
		page = rod.New().ControlURL(u).MustConnect().MustPage(asuDiningWebsiteURL)
	} else {
		page = rod.New().MustConnect().MustPage(asuDiningWebsiteURL)
	}

	router := page.HijackRequests()
	defer router.MustStop()

	var wg sync.WaitGroup

	// Intercept responses from menu server
	router.MustAdd("https://asu.campusdish.com/api/menu/GetMenus*", func(req *rod.Hijack) {
		defer wg.Done()

		log.Info("Intercepted request", "url", req.Request.URL().String())
		// get the response
		req.MustLoadResponse()

		body := req.Response.Body()
		dishes, err := r.ParseMenu(ctx, body)
		if err != nil {
			log.Error("Failed to parse menu", "cause", err)
			return
		}

		err = r.TelegramAPI.SendMessage(ctx, personalID, dishes.ToString())
		if err != nil {
			log.Error("Failed to send message", "cause", err)
		}
	})

	wg.Add(1)
	go router.Run()

	// Check for paragraph inside div with .MenuWrapperDaily
	if page.MustElement(menuWrapper).MustHas("p") {
		// Get the text of the paragraph
		paragraphText := page.MustElement(menuWrapper).MustElement("p").MustText()
		// if text contains noMealsAvailableText, we're done
		if paragraphText == noMealsAvailableText {
			log.Info("No meals available")

			_ = r.TelegramAPI.SendMessage(ctx, personalID, "No meals available")
			return nil
		}
	}

	// Click on the Meal Selection Button
	page.MustElement(mealSelectionButton).MustClick()

	// Enter dinner in the input
	page.MustElement(mealInput).MustInput("dinner")
	if err := page.Keyboard.Press(input.Enter); err != nil {
		log.Error("Failed to press enter", "cause", err)
		return err
	}

	// Click on the Done button
	page.MustElement(doneButton).MustClick()
	page.MustWaitStable()

	wg.Wait()

	return nil
}

func (r *RunHandler) ParseMenu(ctx context.Context, body string) (*Dishes, error) {
	log := logger.FromContext(ctx)

	var response api.Response
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		log.Error("Failed to unmarshal response", "cause", err)
		return nil, err
	}

	// Get period ID for selected period
	var periodID string
	for _, period := range response.Menu.MenuPeriods {
		if period.Name == periodName {
			periodID = period.PeriodID
			break
		}
	}

	log.Info("Period ID", "periodID", periodID)

	// Get station IDs for Home Zone and True Balance
	var homeStationID, trueBalanceStationID string
	for _, station := range response.Menu.MenuStations {
		if station.PeriodID != periodID {
			continue
		}

		switch station.Name {
		case api.HomeStation:
			homeStationID = station.StationID
		case api.TrueBalanceStation:
			trueBalanceStationID = station.StationID
		}
	}

	log.Info("Station IDs", "homeStationID", homeStationID, "trueBalanceStationID", trueBalanceStationID)

	// Get products for Home Zone and True Balance
	var homeZoneDishes, trueBalanceDishes []string
	for _, product := range response.Menu.MenuProducts {
		if product.PeriodID != periodID {
			continue
		}
		switch product.StationID {
		case homeStationID:
			homeZoneDishes = append(homeZoneDishes, product.Product.MarketingName)
		case trueBalanceStationID:
			trueBalanceDishes = append(trueBalanceDishes, product.Product.MarketingName)
		}

	}

	return &Dishes{
		HomeZone:    homeZoneDishes,
		TrueBalance: trueBalanceDishes,
	}, nil
}
