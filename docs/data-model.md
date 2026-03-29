# Datenmodell: Multi-Asset Trading Sandbox

## 1. Modellprinzipien

Das Modell ist asset-agnostisch aufgebaut. XRP ist ein Beispiel-Asset, aber alle Kerntabellen arbeiten generisch mit Assets, Maerkten, Orders, Positionen und Strategien.

Wichtige Prinzipien:

- Asset getrennt von Markt behandeln
- Wallet-Bestaende getrennt von Trades und Positionen fuehren
- jede Strategie-Auswertung nachvollziehbar protokollieren
- manuelle und automatische Orders im selben Modell abbilden

## 2. Kernentitaeten

### users

Speichert Benutzerkonten.

Vorschlag:

- id
- email
- password_hash
- display_name
- created_at
- updated_at

### wallets

Ein Demo-Wallet pro Nutzer, spaeter optional mehrere Wallets.

- id
- user_id
- wallet_type
- base_currency
- starting_balance
- created_at
- updated_at

### wallet_balances

Aktuelle Bestaende je Asset innerhalb eines Wallets.

- id
- wallet_id
- asset_id
- available_amount
- locked_amount
- average_entry_price
- updated_at

### assets

Stammdaten einzelner Coins oder Tokens.

- id
- symbol
- name
- asset_type
- is_active
- created_at

Beispiele:

- XRP
- BTC
- ETH

### markets

Handelspaare wie XRP/USDT oder BTC/USDT.

- id
- base_asset_id
- quote_asset_id
- symbol
- exchange_code
- tick_size
- min_order_size
- is_active
- created_at

### price_ticks

Zeitserien fuer Marktpreise.

- id
- market_id
- timeframe
- timestamp
- open
- high
- low
- close
- volume

Hinweis:
Fuer Realtime kann spaeter ein separater Feed-Store oder Cache genutzt werden. Fuer das MVP reicht persistierte Kerzendatenhaltung.

### orders

Speichert alle Order-Anfragen und Ausfuehrungen, egal ob manuell oder durch Strategie erzeugt.

- id
- user_id
- wallet_id
- market_id
- strategy_id nullable
- order_source
- side
- order_type
- status
- requested_quantity
- requested_quote_amount nullable
- executed_quantity
- average_execution_price
- fee_amount
- fee_asset_id
- slippage_bps
- submitted_at
- executed_at nullable
- cancelled_at nullable

Empfohlene Werte:

- order_source: manual, strategy, system
- side: buy, sell
- order_type: market, limit
- status: pending, filled, partial, cancelled, rejected

### trades

Feingranulare Ausfuehrungen je Order, falls spaeter Split-Fills noetig sind.

- id
- order_id
- market_id
- side
- quantity
- price
- fee_amount
- fee_asset_id
- executed_at

Fuer das MVP koennte eine Order zunaechst genau einen Trade erzeugen.

### positions

Aggregierte Sicht auf geoeffnete oder geschlossene Handelspositionen.

- id
- user_id
- wallet_id
- market_id
- strategy_id nullable
- status
- opened_at
- closed_at nullable
- entry_quantity
- entry_price_avg
- exit_quantity
- exit_price_avg
- realized_pnl
- unrealized_pnl nullable

Empfohlene Werte:

- status: open, closed

### strategies

Definition der Bot-Regeln pro Nutzer.

- id
- user_id
- wallet_id
- market_id
- name
- strategy_type
- config_json
- risk_config_json
- status
- created_at
- updated_at
- last_run_at nullable

Empfohlene strategy_type-Werte:

- dip_buy
- take_profit_stop_loss
- dca
- sma_crossover
- rsi_trigger

Empfohlene status-Werte:

- draft
- active
- paused
- archived

### strategy_runs

Protokolliert einzelne Bewertungslaeufe der Strategie-Engine.

- id
- strategy_id
- started_at
- finished_at
- outcome
- decision
- signal_strength nullable
- details_json

Beispiele:

- outcome: executed, skipped, errored
- decision: buy, sell, hold

### backtests

Speichert Backtest-Definitionen und Ergebnisse.

- id
- user_id
- strategy_id nullable
- market_id
- name
- timeframe
- date_from
- date_to
- config_snapshot_json
- result_summary_json
- created_at

### activity_logs

Klartext- und System-Log fuer Nutzeraktionen und Bot-Entscheidungen.

- id
- user_id
- wallet_id nullable
- strategy_id nullable
- order_id nullable
- log_type
- title
- message
- metadata_json nullable
- created_at

Empfohlene Werte:

- log_type: info, warning, trade, strategy, system

## 3. Beziehungen

- user hat viele wallets
- wallet hat viele wallet_balances
- asset hat viele markets als base oder quote
- market hat viele price_ticks
- wallet hat viele orders
- order kann zu genau einer strategy gehoeren
- order kann einen oder mehrere trades haben
- user und wallet haben viele positions
- user hat viele strategies
- strategy hat viele strategy_runs
- strategy oder market kann viele backtests haben
- user hat viele activity_logs

## 4. Beispiel: XRP-Trade

1. Nutzer hat 10.000 USDT im Demo-Wallet.
2. Markt XRP/USDT ist in `markets` vorhanden.
3. Nutzer sendet Market Buy fuer XRP.
4. In `orders` entsteht ein Eintrag mit source `manual`.
5. In `trades` wird die Ausfuehrung gespeichert.
6. `wallet_balances` fuer USDT sinkt, XRP steigt.
7. Eine `position` wird geoeffnet oder vergroessert.
8. Ein `activity_log` dokumentiert die Aktion.

## 5. API-Sicht auf das Modell

Fuer das MVP werden voraussichtlich folgende API-Bereiche benoetigt:

- /auth
- /markets
- /prices
- /wallets
- /orders
- /positions
- /strategies
- /backtests
- /activity

## 6. Offene Entscheidungen

- ob Positionen strikt FIFO, Durchschnittskosten oder Hybrid-Logik nutzen
- ob Gebuehren immer in Quote-Asset simuliert werden
- welche Timeframes fuer price_ticks persistiert werden
- ob Realtime-Daten direkt in die Datenbank geschrieben oder ueber Cache verarbeitet werden
- wie fein die Backtest-Ergebnisse gespeichert werden sollen

## 7. Empfehlung fuer den Start

Fuer das erste Build wuerde ich es bewusst einfach halten:

- ein Demo-Wallet pro Nutzer
- Market Orders zuerst
- eine Ausfuehrung pro Order
- Kerzendaten statt Tick-by-Tick
- Strategiekonfiguration als JSON
- Logs immer mit menschenlesbarer Begruendung

Damit bleibt das System simpel genug fuer ein MVP, aber offen genug fuer spaetere Boersenanbindung, weitere Assets und komplexere Strategien.
