-- name: UpdateUserRed :one
UPDATE users
SET
    updated_at = NOW(),
    is_chirpy_red = True
WHERE id = $1
RETURNING *;
