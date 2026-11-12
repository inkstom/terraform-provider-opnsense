resource "opnsense_captiveportal_template" "example" {
  name           = "guest"
  content_base64 = filebase64("${path.module}/guest-portal.zip") # index.html plus assets
}

# Use it in a zone:
#   template = opnsense_captiveportal_template.example.id
