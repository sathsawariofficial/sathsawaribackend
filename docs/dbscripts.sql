-- Admin Creation Start
INSERT INTO admins (
    id,
    username,
    password,
    created_at,
    updated_at
)
VALUES (
    'admin',
    'twssawari',
    '$2a$10$vBMOgEfCwRJmzoPYqd6YMu.2QVqEFf3ypvb6ynAC3CYpqNFZmiM6i',
    NOW(),
    NOW()
);

-- Admin Creation End