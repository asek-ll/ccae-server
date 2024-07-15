local im = peripheral.find 'inventoryManager'

local config = {}

local function get_items()
    local items = {}
    for slot, item in pairs(im.getItems()) do
        items[tostring(slot)] = {
            name = item['name'],
            count = item['count'],
            maxStackSize = item['maxStackSize'],
        }
    end
    return items
end

local function remove_item_from_player(slots)
    local total_removed = 0
    for _, slot in pairs(slots) do
        local removed = im.removeItemFromPlayer(config['buffer_direction'], 64, slot)
        total_removed = total_removed + removed
    end
    return total_removed
end

local function add_item_to_player(slots)
    local total_added = 0
    for _, slot in pairs(slots) do
        local added = im.addItemToPlayer(config['buffer_direction'], 64, slot)
        total_added = total_added + added
    end
    return total_added
end

local function send_message()
    return {}
end

local function load_config()
    local cfg_load = loadfile 'config.lua'
    if cfg_load ~= nil then
        config = cfg_load()
    end
end

return function(methods, handlers, wsclient)
    load_config()

    methods['getItems'] = get_items
    methods['removeItemFromPlayer'] = remove_item_from_player
    methods['addItemToPlayer'] = add_item_to_player
    methods['sendMessage'] = send_message
    return {
        buffer_name = config['buffer_name'],
    }
end
