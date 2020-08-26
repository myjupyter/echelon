box.cfg{
    listen = 3301,
} 

function bootstrap()
    
    -- Creating User Table
    users = box.schema.space.create('users', { if_not_exists = true })

    -- Creating DB User
    box.schema.user.create('user', { password = 'secret' })
    box.schema.user.grant('user', 'read,write,execute', 'universe')

    users:format({
        {name = 'id',        type = 'unsigned'},
        {name = 'login',     type = 'string'},
        {name = 'email',     type = 'string'},
        {name = 'password',  type = 'string'},
        {name = 'role',      type = 'string'}
    })
    
    seq = box.schema.sequence.create('seq', { if_not_exists = true })

    users:create_index('primary', {
        type = 'hash',
        unique = true,
        parts = {'id'},
        sequence = 'seq'
    })

    users:create_index('by_email', {
        type = 'hash',
        unique = true,
        parts = {'email'}
    })

    users:create_index('by_login', {
        type = 'hash',
        unique = true,
        parts = {'login'}
    })
    
    users:insert{nil, 'admin', 'admin@mail.ru', '$2a$10$xL5IgTeQvkL5V7.9jMnlWegoA8IjqDD8LYWjtdKqsVyuGUunAu4xu', 'admin'}
end

box.once('bootstrap', bootstrap)
