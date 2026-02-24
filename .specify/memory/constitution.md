<!--
Sync Impact Report:
Version: 1.0.0 → 1.0.1 (Clarifications and alignment with ComplyTime Organization Style Guide)
Modified Principles: N/A (no changes to principle content)
Added Sections: N/A
Removed Sections: N/A
Templates requiring updates:
  ✅ plan-template.md - Constitution Check section aligns with principles
  ✅ spec-template.md - No changes needed (generic template)
  ✅ tasks-template.md - No changes needed (generic template)
  ✅ checklist-template.md - No changes needed (generic template)
Follow-up TODOs:
  - RATIFICATION_DATE: Unknown - marked as TODO
  - Containers section: Referenced in style guide but content not provided - pending addition
-->

# ComplyTime Constitution

## Core Principles

### I. Single Source of Truth (Centralized Constants)

Centralize values used in multiple places or that may change over time. Avoid magic strings (e.g., `"active"`, `"https://api..."`) and magic numbers (e.g., `86400`) inline within logic. Move these values into dedicated files (e.g., `internal/consts/consts.go`, `settings.py`).

**Rationale**: Prevents divergence—updating a timeout from 30s to 60s in one file ensures every part of the application updates automatically. Avoids "shotgun surgery"—you should never search and replace a value across 10 different files to make one logical change, reducing the risk of missing instances and introducing bugs.

### II. Simplicity & Isolation

Improve security and maintenance by removing complexity. Ensure the Single Responsibility Principle by keeping functions small and focused, prioritizing isolated small parts that are easily integrated and tested instead of a monolithic and inflexible approach.

**Rationale**: Small, isolated components reduce cognitive load, improve testability, and minimize the blast radius of changes.

### III. Incremental Improvement

We encourage contributors to improve code they interact with while ensuring individual contributions remain focused. If you identify areas for improvement (refactoring, formatting fixes, better naming) not directly related to the main problem you are solving, first clarify the impact on other core principles and propose the changes if they are worth. Otherwise they should be addressed in a separate feature or Pull Request.

**Rationale**: Keeping aesthetic changes separate from logic fixes ensures that PRs remain atomic and easier for maintainers to review.

### IV. Code is Written for Humans First

Code is read more often than it is written. Optimizing for the reader (your future self or a teammate) is more important than optimizing for the writer's speed.

**Implementation**:
- **Explicit Naming**: Variable and function names MUST clearly describe their intent (e.g., use `days_until_expiration` instead of `d`).
- **Avoid "Clever" Code**: Avoid complex one-liners or obscure language features that require deep mental parsing. If the implementation is hard to explain, it is a bad implementation.
- **Self-Documenting**: The code structure itself MUST explain the logic. Comments MUST explain the *why* (business logic/intent), not the *what* (syntax).

**Rationale**: Readable code reduces onboarding time, prevents bugs from misunderstanding, and enables faster debugging.

### V. Do Not Reinvent the Wheel

Leverage existing solutions, but validate their quality and contribute back.

**Implementation**:
- **Research First**: Always research existing open-source libraries or cloud-native solutions before writing custom code.
- **Vet Dependencies**: Evaluate the library's reliability: check its adoption rate, governance model, maintenance frequency (last commit date), and community health.
- **Contribute Upstream**: If an existing library is close to what we need but missing a feature or bug fix, propose a separate PR contributing to it. Prefer sending a Pull Request to the upstream repository over creating a local workaround or a hard fork.

**Rationale**: Using well-maintained libraries reduces our maintenance burden. Contributing back improves the ecosystem for everyone and reduces the technical debt of maintaining internal patches.

### VI. Composability (The Unix Philosophy)

Write programs and functions that do one thing and do it well. Write programs and functions to work together. Our tools MUST be modular. Output from one tool MUST be easily consumable as input for another (e.g., standard JSON/YAML streams).

**Rationale**: Modular tools enable composition, reuse, and integration with external systems. Standard formats ensure interoperability.

### VII. Convention Over Configuration

Decrease the number of decisions a developer or user needs to make. Use sensible defaults. Users SHOULD only need to specify configuration for things that deviate from the standard norm.

**Rationale**: Reduces cognitive load, accelerates onboarding, and prevents configuration errors through sensible defaults.

## Repository Structure & Standards

Every repository under the ComplyTime organization MUST contain the following standard files in the root directory to ensure a consistent developer experience:

| File | Description | Standard |
|:---|:----|:----|
| `README.md` | Project overview, installation, and usage. | Markdown |
| `LICENSE` | Legal terms of use. | **Apache License 2.0** |
| `CONTRIBUTING.md` | Guidelines for contributors. | Link to org-wide guide or repo-specific details. |
| `CODE_OF_CONDUCT.md` | Community standards. | Standard Contributor Covenant |
| `SECURITY.md` | Security policy. | Vulnerability reporting instructions |
| `.github/` | GitHub configuration. | Issue templates, PR templates, workflows. |

Content of all these files are preferably linked to org-wide files, and eventually incremented with repository specific content.

## Contribution Workflow

### Branching Strategy

- **Main Branch**: `main` is the stable production branch.
- **Feature Branches**: Create branches from `main` for all changes.

### Pull Requests (PRs)

- **Atomic Changes**: PRs MUST be small enough to be reviewed in one sitting. Large, sprawling PRs MAY be requested to be split.
- **Review Requirement**: All PRs REQUIRE review from at least two Maintainers.
- **CI/CD Gates**:
  - **Standard**: All PRs MUST generally pass automated checks (linting, testing, build) before merging.
  - **Exceptions**: We recognize that checks MAY occasionally fail due to external issues outside our control or transient flakes that pass locally. In these rare instances, maintainers CAN discuss and agree on exceptions to merge specific PRs despite a red status.
- **Pull Request Title Format**: `<type>: <description>` (e.g., `feat: implement oscal validation logic`)

### Commit Messages

Follow the **Conventional Commits** [specification](https://www.conventionalcommits.org/).

## Infrastructure Standards Centralization

We SHOULD centralize workflows, configurations, and templates as much as possible. Refer to [org-infra](https://github.com/complytime/org-infra).

## Coding Standards

### Guidelines for All Programming Languages

- **Empty Line at End of File**: Ensure that all files include an empty line at the end. This helps with version control diffs and adheres to POSIX standards.
- **Pre-commit Hooks**: The pre-commit and pre-push hooks can be configured by installing [pre-commit](https://pre-commit.com/).
- **Makefile**: Use Makefile to centralize code specific commands.
- **Testing**: Write tests for your code. Use descriptive names for test functions and include edge cases. Always test inputs from external sources. Ensure at least a positive and a negative test case to ensure errors and exceptions covered and properly treated.
- **Line Length**: Limit lines to 99 characters unless in exceptional cases where it is reasonable to improve readability.
- **Lint**: Ensure no lint issues according to the lint settings defined in the repository. No trailing spaces.

### Go (e.g., `complyctl`)

#### General Guidelines

- **File Naming**: Use lowercase letters and underscores for file names (e.g., `my_file.go`).
- **Package Names**: Use short, concise, and lowercase names for packages. Avoid underscores and mixed caps.
- **Error Handling**: Always check for errors and handle them appropriately. Return errors to the caller when necessary.

#### Licensing and File Headers

```go
// SPDX-License-Identifier: Apache-2.0
```

#### Code Formatting

Formatting SHOULD be aligned with native go format tools, [`goimports`](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) and [`go fmt`](https://go.dev/blog/gofmt).

#### Additional Guidelines

Other [Go checks](https://github.com/complytime/complyctl/blob/main/.golangci.yml) are present in CI/CD, and therefore it MAY be useful to also run them locally before submitting a PR.

### Python (e.g., `complyscribe`)

#### General Guidelines

- **Type Hinting**: Use Python type hints to improve readability and tooling support.

#### Licensing and File Headers

```python
# SPDX-License-Identifier: Apache-2.0
```

#### Code Formatting

- **Style**: Uses `black` and `isort` for formatting.
- **Lint**: Use `ruff` for linting.
- **Static type check**: Use `mypy` as static type checker.
- **Non-Python files**: Use [Megalinter](https://github.com/oxsecurity/megalinter) to lint in a CI task. See [megalinter.yaml](https://github.com/complytime/complyscribe/blob/main/.mega-linter.yml) for more information.

### Containers

<!-- TODO(CONTAINERS_GUIDE): Containers guide referenced in style guide but content not yet provided. Add containerization standards when available. -->

## Governance

This constitution supersedes all other practices and serves as the central source of truth for engineering standards, contribution workflows, and architectural principles for the ComplyTime organization. All contributors and maintainers are expected to adhere to these guidelines to ensure consistency, quality, and compliance.

### Amendment Procedure

- Amendments REQUIRE documentation of the rationale and impact assessment.
- Amendments MUST be reviewed by at least two Maintainers.
- Version MUST increment according to semantic versioning:
  - **MAJOR**: Backward incompatible governance/principle removals or redefinitions.
  - **MINOR**: New principle/section added or materially expanded guidance.
  - **PATCH**: Clarifications, wording, typo fixes, non-semantic refinements.

### Compliance Review

- All PRs/reviews MUST verify compliance with this constitution.
- Complexity MUST be justified when deviating from principles.
- Use this constitution for runtime development guidance and decision-making.

**Version**: 1.0.2 | **Ratified**: TODO(RATIFICATION_DATE): Original adoption date unknown | **Last Amended**: 2026-02-24
