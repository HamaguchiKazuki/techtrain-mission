-- manage user table

CREATE TABLE IF NOT EXISTS game_user.users (
    id INT(11) NOT NULL AUTO_INCREMENT,
    user_name VARCHAR(30) NOT NULL,
    PRIMARY KEY(id)
);