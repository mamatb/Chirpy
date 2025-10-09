-- name: GetUserFromRefreshToken :one
SELECT * FROM users
WHERE id = (
    SELECT user_id FROM refresh_tokens
    WHERE token = $1 AND revoked_at IS NULL AND expires_at > NOW()
);
