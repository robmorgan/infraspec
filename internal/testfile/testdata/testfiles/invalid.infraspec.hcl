run "broken" {
  module = "./modules/vpc"

  # Missing closing brace for assert
  assert {
    condition = true
    error_message = "test"

  # Unclosed run block too
