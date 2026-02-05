run "same_name" {
  module = "./modules/a"

  assert {
    condition     = true
    error_message = "first run"
  }
}

run "same_name" {
  module = "./modules/b"

  assert {
    condition     = true
    error_message = "duplicate run name"
  }
}
