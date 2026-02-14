package main

import (
	"context" //lifecycle control
	"encoding/json"
	"fmt"
	"log" //error logging
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Input struct {
	City    string `json:"city" jsonschema:"The name of the city to aquire the weather for"`
	Country string `json:"country" jsonschema:"The name of the country"`
}

type Output struct {
	Temperature   string  `json:"temperature" jsonschema:"The temperature of the city"`
	Humidity      float64 `json:"humidity" jsonschema:"The humidity of the city"`
	Precipitation float64 `json:"precipitation" jsonschema:"The precipitation of the city"`
	Condition     string  `json:"condition" jsonschema:"The weather conditions of the city"`
}

type APIResponse struct {
	Current struct {
		Temp_F    float64 `json:"temp_f"`
		Humidity  float64 `json:"humidity"`
		PrecipIn  float64 `json:"precip_in"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

func getWeatherData(ctx context.Context, req *mcp.CallToolRequest, input Input) (*mcp.CallToolResult, Output, error) {
	key := os.Getenv("WEATHER_API_KEY")
	if key == "" {
		return nil, Output{}, fmt.Errorf("Something has went wrong with API")
	}

	loc := strings.TrimSpace(input.City) + "," + strings.TrimSpace(input.Country) // City,Country

	call := "http://api.weatherapi.com/v1/current.json?key=" + key + "&q=" + url.QueryEscape(loc) + "&aqi=no"

	request, err := http.Get(call)

	if err != nil {
		return nil, Output{}, fmt.Errorf("Error: Could not fetch weather")
	}
	defer request.Body.Close()

	var values APIResponse
	decode := json.NewDecoder(request.Body)
	checker := decode.Decode(&values)
	if checker != nil {
		return nil, Output{}, fmt.Errorf("Error decoding: %w", checker)
	}

	out := Output{
		Temperature:   fmt.Sprint(values.Current.Temp_F) + "°F",
		Humidity:      values.Current.Humidity,
		Precipitation: values.Current.PrecipIn,
		Condition:     values.Current.Condition.Text,
	}

	return nil, out, nil
}

func main() {
	godotenv.Load("api.env") //API Key
	server := mcp.NewServer(&mcp.Implementation{Name: "weather_manager", Version: "v1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{Name: "Get_Weather", Description: "Obtains weather for a city"}, getWeatherData) //Adding tool to MCP Server

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}

}
