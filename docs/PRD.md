# PRD: Multi-Asset Trading Sandbox

## 1. Produktueberblick

Die Multi-Asset Trading Sandbox ist eine Web-App fuer simuliertes Krypto-Trading mit virtuellen Guthaben, automatisierten Strategien und Backtesting. XRP dient als Referenz-Asset fuer das erste Nutzererlebnis, die Produktarchitektur ist jedoch von Beginn an auf mehrere Assets und Maerkte ausgelegt.

Das Produkt verbindet drei Anwendungsfaelle:

- manuelles Paper-Trading mit virtuellem Geld
- automatisiertes Demo-Trading mit Strategien und Bots
- Analyse von Performance, Positionen und Handelsgruenden

Das MVP handelt ausschliesslich im Demo-Modus. Echte Exchange-Anbindungen sind eine spaetere Erweiterung und kein Bestandteil der ersten Version.

## 2. Problem

Viele bestehende Loesungen decken nur einen Teil des Workflows ab:

- reine Preis-Tracking-Apps ohne Handelslogik
- Demo-Trading ohne Automatisierung
- Bot-Tools ohne gute Visualisierung oder Lernkurve
- komplexe Profi-Tools, die fuer Einsteiger schwer zugaenglich sind

Nutzer sollen an einem Ort lernen, testen und optimieren koennen, ohne echtes Kapital zu riskieren.

## 3. Zielgruppen

### Primaere Zielgruppen

- Einsteiger, die Trading mit Spielgeld verstehen moechten
- Nutzer, die automatisierte Strategien ausprobieren wollen
- Trader, die Regeln erst simulieren wollen, bevor sie echtes Geld einsetzen

### Sekundaere Zielgruppen

- Content-Creator, die Strategien demonstrieren wollen
- Communities oder Lerngruppen, die Setups vergleichen moechten

## 4. Produktziele

- Ein glaubwuerdiges Demo-Trading-Erlebnis fuer mehrere Coins bieten
- Bot-Strategien ohne Code fuer normale Nutzer konfigurierbar machen
- nachvollziehbar erklaeren, warum Trades ausgefuehrt wurden
- XRP als schnellen Einstieg nutzen, ohne das System auf XRP zu begrenzen

## 5. Nicht-Ziele fuer das MVP

- keine echte Boersenanbindung
- kein echtes Geld oder Custody
- kein Social Trading
- keine komplexen Derivate oder Margin-Produkte
- kein KI-gesteuertes Trading im MVP

## 6. Kernannahmen

- Nutzer bevorzugen einen schnellen Einstieg mit einem vorausgewaehlten Markt wie XRP/USDT
- Multi-Asset-Unterstuetzung ist fuer Glaubwuerdigkeit und Wiederverwendbarkeit frueh wichtig
- einfache Regeln wie Dip-Buy, Take-Profit und Stop-Loss reichen fuer ein erstes MVP
- Transparenz ueber Handelsentscheidungen ist ein zentrales Vertrauenselement

## 7. MVP-Funktionsumfang

### 7.1 Auth und Nutzerkonto

- Registrierung und Login
- persoenliches Demo-Konto
- Startguthaben, zum Beispiel 10.000 USDT

### 7.2 Marktuebersicht

- Watchlist mit mehreren Coins
- Start mit XRP, BTC, ETH, SOL, ADA
- Kursanzeige und prozentuale Veraenderung
- Such- und Filterfunktion fuer Maerkte

### 7.3 Asset-Detailseite

- Kurschart
- Marktinformationen
- Order-Modul fuer Demo-Kauf und Demo-Verkauf
- Uebersicht ueber offene Positionen und letzte Trades

### 7.4 Demo-Trading

- Market Buy
- Market Sell
- optional spaeter Limit Orders
- simulierte Gebuehren
- einfache Slippage-Logik

### 7.5 Bot-Strategien

- Strategie an Markt koppeln, z. B. XRP/USDT
- Aktivieren und Deaktivieren von Strategien
- erste Strategie-Typen:
  - Dip Buy
  - Take Profit
  - Stop Loss
  - DCA
  - SMA Crossover
  - RSI Trigger

### 7.6 Portfolio

- Gesamtwert des Demo-Portfolios
- Allokation nach Asset
- realisierte und unrealisierte Gewinne/Verluste
- Performance ueber Zeit

### 7.7 Aktivitaets- und Trade-Log

- jede ausgefuehrte Order historisch speichern
- Ursprung der Order ausweisen:
  - manuell
  - Strategie
  - Systemaktion
- Klartext-Begruendung anzeigen, warum ein Bot gehandelt hat

### 7.8 Einfaches Backtesting

- historische Kursdaten fuer einen Markt laden
- Regel auf Zeitraum anwenden
- Kennzahlen anzeigen:
  - Rendite
  - Trefferquote
  - maximaler Drawdown
  - Anzahl der Trades

## 8. Beispielhafte Nutzerfluesse

### Flow A: Einsteiger testet XRP

1. Nutzer registriert sich.
2. Die App oeffnet auf XRP/USDT.
3. Nutzer sieht Kurs, Chart und Demo-Guthaben.
4. Nutzer kauft XRP mit virtuellem Kapital.
5. Portfolio und Position werden aktualisiert.

### Flow B: Nutzer aktiviert einen Bot

1. Nutzer waehlt XRP/USDT oder BTC/USDT.
2. Nutzer erstellt eine Dip-Buy-Strategie.
3. Nutzer setzt Regeln und Positionsgroesse.
4. Strategie wird aktiviert.
5. Der Bot fuehrt bei passenden Bedingungen virtuelle Orders aus.
6. Die App zeigt Log und Performance der Strategie.

### Flow C: Nutzer testet eine Regel historisch

1. Nutzer waehlt einen Markt und einen Zeitraum.
2. Nutzer waehlt eine Strategie.
3. Das Backtesting berechnet simulierte Trades.
4. Nutzer vergleicht Ergebnis und Risiko.

## 9. Funktionsanforderungen

### Muss-Anforderungen

- mehrere Assets und Maerkte unterstuetzen
- Demo-Wallet pro Nutzer fuehren
- manuelle Demo-Orders ausfuehren
- Strategien speichern und evaluieren
- Trades und Positionen nachvollziehbar speichern
- Performance und Log-Eintraege anzeigen

### Soll-Anforderungen

- einfache Backtests
- Realtime- oder Near-Realtime-Kursdaten
- Watchlist und Suchfunktion

### Kann-Anforderungen

- Notifications
- Strategie-Vorlagen
- Vergleich mehrerer Strategien

## 10. Qualitaetsanforderungen

- Nachvollziehbarkeit: jede Systementscheidung muss auditierbar sein
- Erweiterbarkeit: neue Assets und Strategie-Typen ohne Umbau des Kerns
- Glaubwuerdigkeit: Gebuehren und Slippage nicht ignorieren
- Sicherheit: keine echten API-Keys im MVP erforderlich
- Bedienbarkeit: Einsteiger sollen in wenigen Minuten die erste Demo-Order ausfuehren

## 11. KPIs fuer das MVP

- Anteil der Nutzer, die ihre erste Demo-Order ausfuehren
- Anteil der Nutzer, die mindestens eine Strategie aktivieren
- durchschnittliche Anzahl aktivierter Strategien pro Nutzer
- Retention nach 7 Tagen
- Anzahl durchgefuehrter Demo-Trades pro aktivem Nutzer

## 12. Risiken

- schlechte oder unstete Marktdaten fuehren zu unplausiblen Simulationen
- zu komplexe Strategie-Konfiguration ueberfordert Einsteiger
- unrealistische Demo-Ausfuehrung verfremdet spaetere Echtgeld-Erwartungen
- fehlende Transparenz bei Bot-Entscheidungen schwaecht Vertrauen

## 13. Roadmap nach MVP

### Phase 2

- Limit Orders
- verbesserte Backtests
- Benachrichtigungen
- Strategie-Templates

### Phase 3

- Paper-Trading auf mehreren simulierten Accounts
- fortgeschrittenes Risk Management
- Vergleichs-Reports je Strategie

### Phase 4

- optionale echte Exchange-Anbindung
- API-Key-Management
- Live-Trading nur mit klaren Sicherheitsmechanismen
