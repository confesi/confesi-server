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


----------------------------------- For non-roots

SELECT c.*, c.children_count
FROM "comments" c
WHERE ARRAY_LENGTH(ancestors, 1) = 1 -- num of parents above including root
AND ancestors[1] = 34 -- num of parents above including root -- parent id
ORDER BY updated_at
LIMIT 3;

----------------------------------- For roots

SELECT c.*, c.children_count
FROM "comments" c
WHERE COALESCE(ancestors, '{}') = '{}'
ORDER BY score
LIMIT 3;

-----------------------------------

SELECT *
FROM "comments"
WHERE ARRAY_LENGTH(ancestors, 1) = 1
AND ancestors[1] = 34
ORDER BY score DESC

-----------------------------------

WITH RECURSIVE comment_hierarchy AS (
  -- Anchor query: Fetch the root comments
  SELECT *
  FROM comments
  WHERE ancestors[1] = 34
  
  UNION
  
  -- Recursive query: Fetch the child comments for each parent
  SELECT c.*
  FROM comments c
  INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors)
  WHERE ARRAY_LENGTH(c.ancestors, 1) = ARRAY_LENGTH(ch.ancestors, 1) + 1
)
SELECT *
FROM comment_hierarchy
ORDER BY score DESC
LIMIT 2;


-------------------------------------fetch non-roots (ie: thread)

WITH RECURSIVE comment_hierarchy AS (
  -- Anchor query: Fetch the root comments
  SELECT *
  FROM comments
  WHERE ancestors[1] = 34
  OR COALESCE(ancestors, '{}') = '{}'

  
  UNION
  
  -- Recursive query: Fetch the child comments for each parent
  SELECT c.*
  FROM comments c
  INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors)
  WHERE ARRAY_LENGTH(c.ancestors, 1) = ARRAY_LENGTH(ch.ancestors, 1) + 1
)
SELECT *
FROM comment_hierarchy
ORDER BY score DESC
LIMIT 10;

----------------------------------- fetch roots and thread (wip)

SELECT *
FROM (
	WITH RECURSIVE comment_hierarchy AS (
  -- Anchor query: Fetch the root comments
  SELECT *
  FROM comments
--   WHERE ancestors[1] = 34
  
  UNION
  
  -- Recursive query: Fetch the child comments for each parent
  SELECT c.*
  FROM comments c
  INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors)
  WHERE ARRAY_LENGTH(c.ancestors, 1) = ARRAY_LENGTH(ch.ancestors, 1) + 1
)
SELECT *
FROM comment_hierarchy
ORDER BY score DESC
LIMIT 10
) as sub
WHERE ancestors[1] = 34

----------------------------------- more progress i guess 

-- todo: "not in already-seen ids" from redis

WITH RECURSIVE comment_hierarchy AS (
  -- Anchor query: Fetch the root comments
  SELECT *
  FROM comments
  WHERE ancestors[1] in (
  	SELECT comments.id
  	FROM comments
  	WHERE COALESCE(ancestors, '{}') = '{}'
  	AND comments.post_id = 1 -- [input] some post id
  	ORDER BY score -- root sort
  	LIMIT 3 -- [input] limit of roots to return
  )
  
  UNION
  
  -- Recursive query: Fetch the child comments for each parent
  SELECT c.*
  FROM comments c
  INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors)
  WHERE ARRAY_LENGTH(c.ancestors, 1) = ARRAY_LENGTH(ch.ancestors, 1) + 1
)
SELECT *
FROM comment_hierarchy
ORDER BY score DESC
LIMIT 5 -- [input] limit of children to load per all roots?


-------- CORRECT?!?!?!?!

WITH RECURSIVE comment_hierarchy AS (
  SELECT *,
         1 AS level
  FROM comments
  WHERE ancestors[1] IN (
    SELECT id
    FROM comments
    WHERE COALESCE(ancestors, '{}') = '{}' AND post_id = 1 -- [input] post id
  )
  
  UNION ALL
  
  SELECT c.*,
         ch.level + 1 AS level
  FROM comments c
  INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors)
  WHERE ARRAY_LENGTH(c.ancestors, 1) = ARRAY_LENGTH(ch.ancestors, 1) + 1
)
SELECT DISTINCT subquery.id, subquery.score, subquery.content
FROM (
  SELECT ch.*, ROW_NUMBER() OVER (PARTITION BY ch.ancestors[1] ORDER BY ch.id) AS rn
  FROM comment_hierarchy ch
  INNER JOIN comments c ON ch.ancestors[1] = c.id
) subquery
WHERE rn <= 1 -- [input] maximum number of children per root comment
ORDER BY score DESC;


----


WITH RECURSIVE comment_hierarchy AS (
  SELECT *,
         1 AS level,
         ROW_NUMBER() OVER (PARTITION BY ancestors[1] ORDER BY score DESC) AS rn
  FROM comments
  WHERE ancestors[1] IN (
    SELECT id
    FROM comments
    WHERE COALESCE(ancestors, '{}') = '{}' AND post_id = 1 -- [input] post id
    ORDER BY score DESC
  )
  
  UNION ALL
  
  SELECT c.*,
         ch.level + 1 AS level,
         ROW_NUMBER() OVER (PARTITION BY c.ancestors[1] ORDER BY c.score DESC) AS rn
  FROM comments c
  INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors)
  WHERE ARRAY_LENGTH(c.ancestors, 1) = ARRAY_LENGTH(ch.ancestors, 1) + 1
)
SELECT DISTINCT id, score, content
FROM comment_hierarchy
WHERE rn <= 2 -- [input] maximum number of children per root comment

------- good v2

WITH RECURSIVE comment_hierarchy AS (
  SELECT c.id, c.score, c.content, ARRAY[c.id] AS path, 1 AS level, c.id AS root_id, 1 AS discovery_order, c.ancestors
  FROM comments c
  WHERE COALESCE(ancestors, '{}') = '{}' AND post_id = 1 -- [input] post id
  ORDER BY c.score
  LIMIT 1

  UNION ALL

  SELECT c.id, c.score, c.content, ch.path || c.id, ch.level + 1, ch.root_id, ch.discovery_order + 1, c.ancestors
  FROM comments c
  INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors) OR (ch.id = c.ancestors[1] AND c.ancestors[1] = ANY(ch.ancestors))
  WHERE ch.level < 3 -- Maximum recursion depth of 3
  AND c.score < ch.score
)
SELECT id, score, content, discovery_order, ancestors
FROM comment_hierarchy
ORDER BY cardinality(ancestors), score DESC, discovery_order
LIMIT 4;

------- good v3

WITH RECURSIVE comment_hierarchy AS (
  SELECT c.id, c.score, c.content, ARRAY[c.id] AS path, 1 AS level, c.id AS root_id, 1 AS discovery_order, c.ancestors
  FROM comments c
  WHERE COALESCE(ancestors, '{}') = '{}' AND post_id = 1 -- [input] post id

  UNION ALL

  SELECT c.id, c.score, c.content, ch.path || c.id, ch.level + 1, ch.root_id, ch.discovery_order + 1, c.ancestors
  FROM comments c
  INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors) OR (ch.id = c.ancestors[1] AND c.ancestors[1] = ANY(ch.ancestors))
  WHERE ch.level < 3 -- Maximum recursion depth of 3
  AND c.score < ch.score
), ranked_comments AS (
  SELECT id, score, content, discovery_order, ancestors, ROW_NUMBER() OVER (PARTITION BY root_id ORDER BY cardinality(ancestors), score DESC, discovery_order) AS row_num
  FROM comment_hierarchy
)
SELECT id, score, content, discovery_order, ancestors
FROM ranked_comments
WHERE row_num <= 4
ORDER BY row_num;


---- meh 

WITH RECURSIVE comment_hierarchy AS (
  SELECT c.id, c.score, c.content, ARRAY[c.id] AS path, 1 AS level, c.id AS root_id, 1 AS discovery_order, c.ancestors
  FROM comments c
  WHERE COALESCE(ancestors, '{}') = '{}' AND post_id = 1 -- [input] post id

  UNION ALL

  SELECT c.id, c.score, c.content, ch.path || c.id, ch.level + 1, ch.root_id, ch.discovery_order + 1, c.ancestors
  FROM comments c
  INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors) OR (ch.id = c.ancestors[1] AND c.ancestors[1] = ANY(ch.ancestors))
  WHERE ch.level < 3 -- Maximum recursion depth of 3
  AND c.score < ch.score
), ranked_comments AS (
  SELECT id, score, content, discovery_order, ancestors, root_id, ROW_NUMBER() OVER (PARTITION BY root_id ORDER BY cardinality(ancestors), score DESC, discovery_order) AS row_num
  FROM comment_hierarchy
)
SELECT id, score, content, discovery_order, ancestors
FROM ranked_comments
WHERE row_num <= 100
ORDER BY root_id, row_num;


----- DONE

WITH RECURSIVE comment_hierarchy AS (
  SELECT c.id, c.score, c.content, ARRAY[c.id] AS path, 1 AS level, c.id AS root_id, 1 AS discovery_order, c.ancestors
  FROM comments c
  WHERE COALESCE(ancestors, '{}') = '{}' AND post_id = 1 -- [input] post id

  UNION ALL

  SELECT c.id, c.score, c.content, ch.path || c.id, ch.level + 1, CASE WHEN ch.root_id = c.id THEN c.id ELSE ch.root_id END, ch.discovery_order + 1, c.ancestors
  FROM comments c
  INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors) OR (ch.id = ch.root_id AND ch.root_id != c.id)
  WHERE ch.level < 5 -- Maximum recursion depth of 5
  AND c.score < ch.score
), ranked_comments AS (
  SELECT id, score, content, discovery_order, ancestors, root_id, ROW_NUMBER() OVER (PARTITION BY root_id ORDER BY cardinality(ancestors), score DESC, discovery_order) AS row_num
  FROM comment_hierarchy
)
SELECT id, score, content, discovery_order, ancestors
FROM ranked_comments
WHERE row_num <= 100
ORDER BY root_id, row_num;


------- OOF

WITH RECURSIVE comment_hierarchy AS (
  SELECT c.id, c.score, c.content, ARRAY[c.id] AS path, 1 AS level, c.id AS root_id, 1 AS discovery_order, c.ancestors
  FROM comments c
  WHERE COALESCE(ancestors, '{}') = '{}' AND post_id = 1 -- [input] post id

  UNION ALL

  SELECT c.id, c.score, c.content, ch.path || c.id, ch.level + 1, CASE WHEN ch.root_id = c.id THEN c.id ELSE ch.root_id END, ch.discovery_order + 1, c.ancestors
  FROM comments c
  INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors) OR (ch.id = ch.root_id AND ch.root_id != c.id)
  WHERE ch.level < 5 -- Maximum recursion depth of 5
  AND c.score < ch.score
), ranked_comments AS (
  SELECT id, score, content, discovery_order, ancestors, root_id, ROW_NUMBER() OVER (PARTITION BY root_id ORDER BY cardinality(ancestors), score DESC, discovery_order) AS row_num
  FROM comment_hierarchy
)
SELECT id, score, content, discovery_order, ancestors
FROM ranked_comments
WHERE row_num <= 100
ORDER BY root_id, row_num;

--- perfectly for 1 root without limit per depth

WITH RECURSIVE comment_hierarchy AS (
  SELECT c.id, c.score, c.content, ARRAY[c.id] AS path, 1 AS level,
    c.id AS root_id, 1 AS discovery_order, c.ancestors
  FROM comments c
  WHERE COALESCE(ancestors, '{}') = '{}'
    AND post_id = 1 -- [input] post id
    AND c.id = 40 -- Specify the root comment ID here

  UNION ALL

  SELECT c.id, c.score, c.content, ch.path || c.id, ch.level + 1,
    CASE WHEN ch.root_id = c.id THEN c.id ELSE ch.root_id END,
    ch.discovery_order + 1, c.ancestors
  FROM comments c
  INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors)
    OR (ch.id = ch.root_id AND ch.root_id != c.id)
  WHERE c.score < ch.score
    AND c.ancestors[1] = 40 -- Specify the root comment ID here
), ranked_comments AS (
  SELECT id, score, content, discovery_order, ancestors, root_id,
    ROW_NUMBER() OVER (ORDER BY cardinality(ancestors), score DESC, discovery_order) AS row_num
  FROM comment_hierarchy
), filtered_comments AS (
  SELECT DISTINCT ON (root_id, id) id, score, content, ancestors, row_num, root_id
  FROM ranked_comments
  WHERE row_num <= 10
  ORDER BY root_id, id, row_num
)
SELECT id, score, content, ancestors, row_num, root_id
FROM filtered_comments
ORDER BY row_num;


----

WITH RECURSIVE comment_hierarchy AS (
  SELECT c.id, c.score, c.content, ARRAY[c.id] AS path, 1 AS level,
    c.id AS root_id, 1 AS discovery_order, c.ancestors
  FROM comments c
  WHERE COALESCE(ancestors, '{}') = '{}'
    AND post_id = 1 -- [input] post id
    AND c.id = 40 -- Specify the root comment ID here

  UNION ALL

  SELECT c.id, c.score, c.content, ch.path || c.id, ch.level + 1,
    CASE WHEN ch.root_id = c.id THEN c.id ELSE ch.root_id END,
    ch.discovery_order + 1, c.ancestors
  FROM comments c
  INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors)
    OR (ch.id = ch.root_id AND ch.root_id != c.id)
  WHERE c.score < ch.score
    AND c.ancestors[1] = 40 -- Specify the root comment ID here
), ranked_comments AS (
  SELECT id, score, content, discovery_order, ancestors, root_id,
    ROW_NUMBER() OVER (PARTITION BY path[1:level] ORDER BY cardinality(ancestors), score DESC, discovery_order) AS row_num,
    path
  FROM comment_hierarchy
), filtered_comments AS (
  SELECT id, score, content, ancestors, row_num, root_id, path, level
  FROM (
    SELECT *,
      ROW_NUMBER() OVER (PARTITION BY path[1:cardinality(path)] ORDER BY score DESC, discovery_order) AS depth_row_num,
      cardinality(path) AS level
    FROM ranked_comments
  ) AS subquery
  WHERE depth_row_num <= 2 -- Maximum of 2 comments per depth level per traversal route
), filtered_comments_with_row_num AS (
  SELECT id, score, content, ancestors, row_num, root_id, path, level,
    ROW_NUMBER() OVER (PARTITION BY path[1:level] ORDER BY row_num) AS depth_row_num
  FROM filtered_comments
  WHERE row_num <= 2 -- Additional filter to limit 2 comments per depth level
)
SELECT id, score, content, ancestors, row_num, root_id
FROM filtered_comments_with_row_num
WHERE depth_row_num <= 2 -- Maximum of 2 comments per depth level per traversal route
ORDER BY row_num;


------------

WITH RECURSIVE comment_hierarchy AS (
  SELECT
    c.id,
    c.score,
    c.content,
    ARRAY[c.id] AS path,
    1 AS level,
    c.id AS root_id,
    1 AS discovery_order,
    c.ancestors
  FROM
    comments c
  WHERE
    COALESCE(ancestors, '{}') = '{}'
    AND post_id = 1 -- [input] post id
    AND c.id = 40 -- Specify the root comment ID here

  UNION ALL

  SELECT
    c.id,
    c.score,
    c.content,
    ch.path || c.id,
    ch.level + 1,
    CASE WHEN ch.root_id = c.id THEN c.id ELSE ch.root_id END,
    ch.discovery_order + 1,
    c.ancestors
  FROM
    comments c
    INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors)
      OR (ch.id = ch.root_id AND ch.root_id != c.id)
  WHERE
    c.score < ch.score
    AND c.ancestors[1] = 40 -- Specify the root comment ID here
), ranked_comments AS (
  SELECT
    id,
    score,
    content,
    discovery_order,
    ancestors,
    root_id,
    ROW_NUMBER() OVER (PARTITION BY path[1:level] ORDER BY cardinality(ancestors), score DESC, discovery_order) AS row_num,
    path
  FROM
    comment_hierarchy
), filtered_comments AS (
  SELECT
    id,
    score,
    content,
    ancestors,
    row_num,
    root_id,
    path,
    level
  FROM (
    SELECT
      *,
      ROW_NUMBER() OVER (PARTITION BY path[1:cardinality(path)] ORDER BY score DESC, discovery_order) AS depth_row_num,
      cardinality(path) AS level
    FROM
      ranked_comments
  ) AS subquery
  WHERE
    depth_row_num <= 2 -- Maximum of 2 comments per depth level per traversal route
), paginated_comments AS (
  SELECT
    *,
    ROW_NUMBER() OVER (ORDER BY level, row_num) AS bfs_row_num
  FROM (
    SELECT
      *,
      ROW_NUMBER() OVER (PARTITION BY path[1:cardinality(path) - 1] ORDER BY row_num) AS parent_row_num
    FROM
      filtered_comments
  ) AS subquery
  WHERE
    parent_row_num <= 1 -- Maximum of 2 siblings per parent
)
SELECT
  id,
  score,
  content,
  ancestors,
  row_num,
  root_id
FROM
  paginated_comments
WHERE
  bfs_row_num <= 5 -- Specify the desired pagination range here
ORDER BY
  bfs_row_num;


----- pure bfs

WITH RECURSIVE bfs_comments AS (
  -- Anchor member
  SELECT
    id,
    score,
    0 AS level,
    ARRAY[id] AS path,
    0::bigint AS sibling_order,
    children_count,
    ancestors,
    content
  FROM
    comments
  WHERE
    COALESCE(ancestors, '{}') = '{}'

  UNION ALL

  -- Recursive member
  SELECT
    c.id,
    c.score,
    bc.level + 1,
    bc.path || c.id,
    ROW_NUMBER() OVER (PARTITION BY bc.path ORDER BY c.id)::bigint AS sibling_order,
    c.children_count,
    c.ancestors,
    c.content
  FROM
    comments c
  INNER JOIN bfs_comments bc ON c.ancestors[1] = bc.id
)
SELECT
  id,
  score,
  level,
  path,
  sibling_order,
  children_count,
  ancestors,
  content
FROM
  bfs_comments
ORDER BY
  path;
