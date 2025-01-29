Feature: Terraform Hello World

  Scenario: Run a simple test
    Given the Terraform module at "./examples/terraform-hello-world-example"
    When I run "terraform apply"
    Then the "hello_world" output is "Hello, World!"
