Feature: Terraform Hello World

  Scenario: Run a simple test
    Given the Terraform module at "./fixtures/hello-world"
    When I run terraform apply
    Then the "hello_world" output is "Hello, World!"
