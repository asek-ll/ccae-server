local im = peripheral.find 'inventoryManager'

local storage_direction = 'east'
-- local function is_storage(peripheral_name)
--     local methods = peripheral.getMethods(peripheral_name)
--     local has_get_item_details = false
--     for _, method in pairs(methods) do
--         print(method)
--         if method == 'getItemDetail' then
--             has_get_item_details = true
--         end
--     end
--     return has_get_item_details
-- end
-- for _, v in pairs(peripheral.getNames()) do
--     if is_storage(v) then
--         storage_direction = v
--     end
-- end

print('Detected storage' .. storage_direction)

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

local function remove_item_from_player(slot)
    return im.removeItemFromPlayer(storage_direction, 64, slot)
end

return function(methods, handlers, wsclient)
    methods['getItems'] = get_items
    methods['removeItemFromPlayer'] = remove_item_from_player
end
