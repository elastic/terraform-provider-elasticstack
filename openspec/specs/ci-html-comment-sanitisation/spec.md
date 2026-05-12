# ci-html-comment-sanitisation Specification

## Purpose
Shared deterministic helpers for stripping HTML comments from untrusted human-authored content before it is passed to agentic workflows, eliminating a prompt-injection surface where malicious users embed fake markers or directives inside HTML comment syntax.

## Requirements

### Requirement: Shared library provides HTML-comment stripping
The repository SHALL provide a shared helper under `.github/workflows-src/lib/` that exposes a `stripHtmlComments(text)` function. The function SHALL remove all HTML comment sequences — that is, text matching the pattern `<!--` through the next `-->` — from an input string and SHALL return the remaining text. When an opening `<!--` is present without a matching closing `-->`, the function SHALL remove from that `<!--` to the end of the string. The helper SHALL be unit-testable independently of any workflow wrapper.

#### Scenario: Input contains embedded HTML comment
- **WHEN** `stripHtmlComments` is called with text containing `<!-- hidden -->`
- **THEN** the returned string SHALL NOT contain `<!-- hidden -->` or any portion of it

#### Scenario: Input contains multiple HTML comments
- **WHEN** `stripHtmlComments` is called with text containing several disjoint HTML comments
- **THEN** the returned string SHALL have all of them removed

#### Scenario: Input contains no HTML comments
- **WHEN** `stripHtmlComments` is called with plain text lacking any `<!--` sequence
- **THEN** the returned string SHALL be identical to the input

#### Scenario: Input contains an unterminated HTML comment
- **WHEN** `stripHtmlComments` is called with text containing `<!-- hidden` with no closing `-->`
- **THEN** the returned string SHALL remove from the `<!--` through the end of the string
- **AND** the injection surface SHALL be eliminated

#### Scenario: Maintainer inspects shared library
- **WHEN** maintainers inspect `.github/workflows-src/lib/` for sanitisation helpers
- **THEN** they SHALL find a module exporting `stripHtmlComments` with focused unit tests

### Requirement: Factory workflows strip HTML comments before agent context
Every factory workflow that feeds a GitHub issue body or human-authored issue comments into an agent prompt SHALL apply `stripHtmlComments` to that content in deterministic pre-activation steps before writing the context files. The sanitisation SHALL apply to `research-factory`, `change-factory`, and `code-factory` workflows. Bot-authored comments — including the research-factory sticky comment itself — SHALL NOT be passed through this stripping step because they are trusted output from prior runs.

#### Scenario: Research-factory sanitises issue body
- **WHEN** the `research-factory` workflow prepares the `issue_body.md` context file
- **THEN** it SHALL run `stripHtmlComments` on the issue body text first
- **AND** the agent SHALL receive only the sanitised body

#### Scenario: Research-factory sanitises human comment history
- **WHEN** the `research-factory` workflow prepares the `issue_comments.md` context file
- **THEN** it SHALL run `stripHtmlComments` on each human-authored comment before inclusion
- **AND** the agent SHALL receive only sanitised comment text

#### Scenario: Change-factory sanitises issue body and human comments
- **WHEN** the `change-factory` workflow reads the triggering issue body and comments
- **THEN** it SHALL apply `stripHtmlComments` to the issue body and to human-authored comments
- **AND** the agent SHALL receive only sanitised input

#### Scenario: Code-factory sanitises issue body and human comments
- **WHEN** the `code-factory` workflow reads the triggering issue body and comments
- **THEN** it SHALL apply `stripHtmlComments` to the issue body and to human-authored comments
- **AND** the agent SHALL receive only sanitised input

#### Scenario: Research comment is not stripped
- **WHEN** a `change-factory` or `research-factory` workflow reads a bot-authored comment containing `<!-- gha-research-factory -->`
- **THEN** that comment SHALL be passed to the agent verbatim without HTML-comment stripping
- **AND** the agent SHALL see the intact marker for extraction or reference
