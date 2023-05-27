# some waacky stuff

### TODO

- look into caesar cipher with golang -> mask userid
  - probs need a middleware and pass down the context (user id)

## auth

- user-email auth.
- sends cred to server, verify school before registration.
- anon option for read-only.

## api middleware

- JWT
  - X-Token for App Check. <- pub/sub key
  - account_state: anon | register

## DB

- Firebase or SQL.

## Tables

### ~~`Posts` table~~

- created_by
- metadata: created by / updated by / etc
- school_id
- faculty_id: nullable
- title: could be null - handled client side
- downvote: number
- upvote: number
- score?
- trending_score: number <- matt will deal with this
- hottest_on: date | null
- hidden: boolean

### ~~`Users`~~

- meta
- email
- password <- no need for this
- school_id
- year of study: nullable
- faculty_id
- is_banned: boolean
- mod_id
- role <- deferred feat

  - server verifies email for school's domain email -> checks for school

### ~~`Schools`~~

- id
- name
- abbr
- lat: float
- lon: float
  ie: 40.753, -73.983.
- domain: string

### ~~`Faculty`~~

- id
- faculty

### ~~`mod_level`~~

- enable
- ban
- limited
- degen

### ~~school_follows~~

- id
- user_id
- school_id

### ~~votes~~

- id
- user_id
- comment_id
- post_id
- vote: "1" || "-1 <- in" int though

### ~~comments~~

- id
- user_id
- post_id
- comment_id
- content
- upvote <- count
- downvote <- count
- score
- hidden: boolean

### ~~saved_posts~~

- id
- meta
- post_id

### ~~saved_comments~~

- id
- meta
- comment_id

### ~~feedbacks~~

- meta
- user_id
- content: text
- type <- deferred feature

### ~~reports~~

- meta
- id
- user_id
- type
- description
- resolution: text
- user_alerted: bool
