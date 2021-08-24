CREATE  TABLE IF NOT EXISTS authors (
                                        id                   serial   PRIMARY KEY ,
                                        "position"           smallint DEFAULT 0 NOT NULL ,
                                        logo                 varchar  NOT NULL ,
                                        name                 varchar  NOT NULL ,
                                        logo_mini            varchar  NOT NULL ,
                                        email                varchar  NOT NULL
);

CREATE  TABLE IF NOT EXISTS authors_text (
                                        id                   serial   PRIMARY KEY ,
                                        author_id            integer  REFERENCES authors (id) NOT NULL ,
                                        hl                   varchar  NOT NULL ,
                                        title                varchar  NOT NULL ,
                                        description          text
);

CREATE  TABLE IF NOT EXISTS categories (
                                      id                   serial  PRIMARY KEY ,
                                      "position"           smallint DEFAULT 0 NOT NULL
);

CREATE  TABLE IF NOT EXISTS categories_text (
                                           id                   serial  PRIMARY KEY ,
                                           category_id          integer  REFERENCES categories (id) NOT NULL ,
                                           hl                   varchar  NOT NULL ,
                                           title                varchar  NOT NULL
);

CREATE  TABLE IF NOT EXISTS news (
                                id                   serial  PRIMARY KEY ,
                                category_id          integer REFERENCES categories (id) DEFAULT 0 NOT NULL ,
                                author_id            integer REFERENCES authors (id) NOT NULL ,
                                image                varchar  NOT NULL ,
                                "date"               timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL ,
                                publish_date         timestamp NOT NULL ,
                                view_count           integer DEFAULT 0 NOT NULL ,
                                open_count           integer DEFAULT 0 NOT NULL ,
                                big                  smallint DEFAULT 0 NOT NULL ,
                                status               integer DEFAULT 0 NOT NULL ,
                                last_upd             integer  DEFAULT 0 NOT NULL
);

CREATE  TABLE IF NOT EXISTS news_text (
                                     id                   serial  PRIMARY KEY ,
                                     news_id              integer  REFERENCES news (id) NOT NULL ,
                                     hl                   varchar  NOT NULL ,
                                     title                varchar  NOT NULL ,
                                     url                  varchar  NOT NULL ,
                                     original             varchar
);

CREATE  TABLE IF NOT EXISTS tags (
                       id                   serial  PRIMARY KEY
);

CREATE  TABLE IF NOT EXISTS tags_text (
                            id                   serial  PRIMARY KEY ,
                            name                 varchar  NOT NULL ,
                            tag_id               integer  REFERENCES tags (id) NOT NULL ,
                            hl                   varchar  NOT NULL
);

CREATE  TABLE IF NOT EXISTS users (
                        id                   serial  PRIMARY KEY ,
                        login                varchar  NOT NULL ,
                        name                 varchar   ,
                        created              timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL ,
                        "rule"               smallint DEFAULT 0 NOT NULL
);

CREATE  TABLE IF NOT EXISTS news_content (
                               id                   serial  PRIMARY KEY ,
                               "value"              varchar  NOT NULL ,
                               tag                  varchar  NOT NULL ,
                               attr                 json  ,
                               news_text_id         integer  REFERENCES news_text( id ) NOT NULL
);

CREATE  TABLE IF NOT EXISTS news_tags (
                            id                   serial  PRIMARY KEY,
                            news_id              integer REFERENCES news( id ) ON DELETE CASCADE NOT NULL ,
                            tag_id               integer REFERENCES tags( id ) ON DELETE CASCADE NOT NULL
);
