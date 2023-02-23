table "notes" {
  schema = schema.public
  column "id" {
    null = false
    type = character_varying(64)
  }
  column "title" {
    null = false
    type = character_varying(1024)
  }
  column "content" {
    null = true
    type = text
  }
  primary_key {
    columns = [column.id]
  }
}
schema "public" {
}
