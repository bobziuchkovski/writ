# Writ Changelog

## 0.8.6 (2016-01-27)
- Fix: Update exported field check for Go 1.6
- Docs: Misc updates and clarifications

## 0.8.5 (2016-01-24)

- Fix: Add a missing nil check to NewOptionDecoder
- Fix: Fix wrapping for multi-line descriptions
- Tests: Add coverage for remaining code, except Command.ExitHelp().  Coverage is at 98.7%.
- Docs: Overhaul docs and examples for brevity
- Docs: Add an example for subcommand handling

## 0.8.4 (2016-01-22)

- Feature: Hide options and commands with empty descriptions from help output

## 0.8.3 (2016-01-22)

- Misc: Minor code cleanup
- Tests: Add basic test coverage for help output
- Tests: Add additional test coverage for comamnds and options

## 0.8.2 (2016-01-22)

- Fix: Stop parsing subcommands after a bare "-" argument
- Fix: Ensure command and option names have no spaces in them
- Tests: Add additional test coverage for comamnds and options

## 0.8.1 (2016-01-22)

- API: Panic NewOptionDecoder() if input type is unsupported
- Docs: Add an example of explicitly creating a Command and Options
- Docs: Update documentation

## 0.8.0 (2016-01-22)

- Misc: Initial release on Github
