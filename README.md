# InfraSpec

<div align="center">
<h1 align="center">
<img src="docs/infraspec_logo.png" width="600" />
</h1>
<h3>Write infrastructure tests in plain English, without writing a single line of code.</h3>
</div>

:warning: This project is still under heavy development and probably won't work!

Under the hood, InfraSpec uses [Gherkin](https://cucumber.io/docs/gherkin/) to parse
Go Dog and testing modules from Terratest.

## Why?

Additionally, LLMs are great at generating scenarios using the Gherkin syntax, so you can write tests in plain English
and InfraSpec will translate them into code.

## Installation

```sh
go install github.com/robmorgan/infraspec@latest
```

## Writing Tests

If your using VS Code, we recommend installing the [Cucumber (Gherkin) Full Support](https://marketplace.visualstudio.com/items?itemName=alexkrechik.cucumberautocomplete)
extension for syntax highlighting.
