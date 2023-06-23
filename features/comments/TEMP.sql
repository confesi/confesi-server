SELECT *, ARRAY_LENGTH(ancestors, 1) as depth
FROM "comments"
WHERE 3 = ANY(ancestors)
AND hidden = false
ORDER BY depth;


SELECT *
FROM (
    SELECT *, ARRAY_LENGTH(ancestors, 1) as depth
    FROM "comments"
    WHERE 3 = ANY(ancestors)
    AND hidden = false
) sub
WHERE (depth, id) > (2, 15)
ORDER BY depth, id
LIMIT 5;


SELECT *
FROM (
    SELECT *, ARRAY_LENGTH(ancestors, 1) as depth
    FROM "comments"
    WHERE 3 = ANY(ancestors)
    AND hidden = false
) sub
WHERE (depth, id) > (3, 27)
ORDER BY depth, id
LIMIT 5;

WITH RECURSIVE comment_tree AS (
  SELECT *, ARRAY[comment_id] AS path
  FROM comments
  WHERE comment_id IN (
    SELECT comment_id
    FROM comments
    WHERE ancestors IS NULL
    ORDER BY comment_id
    LIMIT N  -- Specify the number of root comments to fetch
  )
  UNION ALL
  SELECT c.*, ct.path || c.comment_id
  FROM comments c
  JOIN comment_tree ct ON c.comment_id = ANY(ct.ancestors)
)
SELECT *
FROM comment_tree
ORDER BY path;

WITH RECURSIVE comment_tree AS (
  SELECT *, ARRAY[id] AS path
  FROM comments
  WHERE id IN (
    SELECT id
    FROM comments
    WHERE COALESCE(ancestors, '{}') = '{}'
    ORDER BY id
    LIMIT 5  -- Specify the number of root comments to fetch
  )
  UNION ALL
  SELECT c.*, ct.path || c.id
  FROM comments c
  JOIN comment_tree ct ON c.id = ANY(ct.ancestors)
)
SELECT *
FROM comment_tree
WHERE path[1] IN (
  SELECT id
  FROM comment_tree
)
ORDER BY path;



SELECT *
FROM "comments"
WHERE ARRAY_LENGTH(ancestors, 1) <= 2
AND 

SELECT
    id,
    array_length(ancestors, 1) AS depth
FROM
    comments
WHERE
    20 = ANY(ancestors)
ORDER BY
    depth
;

WITH RECURSIVE comment_tree AS (
  SELECT *, ARRAY[id] AS path
  FROM comments
  WHERE id IN (
    SELECT id
    FROM comments
    WHERE ancestors @> ARRAY[24] -- Specify the ancestor ID of the root comment
    ORDER BY id
    LIMIT 5 -- Specify the number of root comments to fetch
  )
  UNION ALL
  SELECT c.*, ct.path || c.id
  FROM comments c
  JOIN comment_tree ct ON c.id = ANY(ct.ancestors)
  WHERE c.id > 21 -- Specify the last comment ID from the previous page as the cursor
)
SELECT
  ct.id,
  ct.content,
  array_length(ct.ancestors, 1) AS depth,
  (SELECT COUNT(*) FROM comments WHERE ancestors @> ct.path AND id != ct.id) AS unreturned_children_count
FROM comment_tree ct
ORDER BY ct.path;


SELECT
    id,
    content,
    array_length(ancestors, 1) AS depth
FROM
    comments
WHERE
    24 = ANY(ancestors)
ORDER BY
    depth
;

SELECT c.*, (c.children_count - (
    SELECT COUNT(*)
    FROM comments
    WHERE ancestors[0] = 34
  )) AS remaining
FROM "comments" c
WHERE ARRAY_LENGTH(ancestors, 1) > 0 AND ancestors[1] = 34
LIMIT 3;


SELECT c.*, c.children_count
FROM "comments" c
WHERE ARRAY_LENGTH(ancestors, 1) = 1 -- num of parents
AND ancestors[1] = 34 -- parent id
ORDER BY updated_at
LIMIT 3;


------- For non-roots

SELECT c.*, c.children_count
FROM "comments" c
WHERE ARRAY_LENGTH(ancestors, 1) = 1 -- num of parents above including root
AND ancestors[1] = 34 -- num of parents above including root -- parent id
ORDER BY updated_at
LIMIT 3;

------ For roots

SELECT c.*, c.children_count
FROM "comments" c
WHERE COALESCE(ancestors, '{}') = '{}'
ORDER BY score
LIMIT 3;