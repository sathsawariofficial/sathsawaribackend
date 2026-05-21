-- Admin Creation Start
INSERT INTO admins (
    id,
    user_name,
    password,
    created_at,
    updated_at
)
VALUES (
    'admin',
    'twssawari',
    'vR7!xK2@pQ9#Lm4$Zw8^Ty1&Nc5*Hs3%Df6!Ba', -- dont run the raw query hash the password
    NOW(),
    NOW()
);

-- Admin Creation End