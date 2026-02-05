run "no_module" {
  # Missing required module attribute

  assert {
    condition     = true
    error_message = "This should fail due to missing module"
  }
}
