# InfraSpec

Write infrastructure tests in plain English, without writing a single line of code.

:warning: This project is still under heavy development and probably won't work!

Under the hood, InfraSpec uses [Gherkin](https://cucumber.io/docs/gherkin/) to parse
Go Dog and testing modules from Terratest.

## Installation

```sh
go install github.com/robmorgan/infraspec@latest
```

## Writing Tests

If your using VS Code, we recommend installing the [Cucumber (Gherkin) Full Support](https://marketplace.visualstudio.com/items?itemName=alexkrechik.cucumberautocomplete) extension for syntax highlighting.
