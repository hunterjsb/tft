# TFT
[![Go](https://github.com/hunterjsb/tft/actions/workflows/go.yml/badge.svg?branch=master&event=push)](https://github.com/hunterjsb/tft/actions/workflows/go.yml)

Discord bot for feeding Teamfight Tactics data into AI agents for post-game and live analysis.

## ⚠️ [WIP] Development
This project is in the early stages of development and currently uses a **development Riot API key** meaning it will expire every **24 hours**. If the key is not refreshed, the tests run during the `go.yml` github action will fail.

## Setup
See `.env.example` for environment variables. They key ones are:
- RIOT_API_KEY
- DISCORD_TOKEN
- OPENAI_API_KEY

## Usage
`go run .` to start the bot.
