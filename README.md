# InfraSpec

BDD tool for testing infrastructure as code.

Write infrastructure tests in plain English, without writing a single line of code.

Under the hood, InfraSpec uses [Gherkin](https://cucumber.io/docs/gherkin/) to parse
Go Dog and testing modules from Terratest.

## Installation

```sh
go install github.com/robmorgan/infraspec@latest
```

## Writing Tests

If your using VS Code, we recommend installing the [Cucumber (Gherkin) Full Support](https://marketplace.visualstudio.com/items?itemName=alexkrechik.cucumberautocomplete) extension for syntax highlighting.
