local m = peripheral.find 'modem' --[[@as ccTweaked.peripherals.WiredModem]]

local config = {}

local function craft(recipe)
    print(recipe)
    local recipe_type = recipe['type']
    local inventory = config['handlers'][recipe_type]
    if inventory == nil then
        return false
    end

    local target_free = m.callRemote(inventory, 'size')
    for _, _ in pairs(m.callRemote(inventory, 'list')) do
        target_free = target_free - 1
    end
    local buffer = config['buffer_name']

    for _, _ in pairs(m.callRemote(buffer, 'list')) do
        target_free = target_free - 1
    end

    if target_free < 0 then
        return false
    end

    for s, _ in pairs(m.callRemote(buffer, 'list')) do
        m.callRemote(buffer, 'pushItems', inventory, s)
    end

    return true
end

local function dump_out()
    return true
end

local function get_types()
    local result = {}

    for type, inv in pairs(config['handlers']) do
        if m.callRemote(inv, 'getItemDetail', 1) == nil then
            table.insert(result, type)
        end
    end

    return result
end

local function setup()
    local cfg_load = loadfile 'config.lua'
    return cfg_load()
end

---@param methods table<string,fun()>
return function(methods, _, _)
    config = setup()
    methods['dumpOut'] = dump_out
    methods['craft'] = craft
    methods['restore'] = craft
    methods['getTypes'] = get_types
    return {
        buffer_name = config['buffer_name'],
    }
end
