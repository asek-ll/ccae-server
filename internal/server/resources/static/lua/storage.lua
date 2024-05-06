local m = peripheral.find 'modem'

local function is_storage(modem, storage_name)
    local methods = modem.getMethodsRemote(storage_name)
    local has_get_item_details = false
    for _, method in pairs(methods) do
        if method == 'getItemDetail' then
            has_get_item_details = true
        end
    end
    return has_get_item_details
end

local function getItems()
    local storages = {}

    for _, storage_name in pairs(m.getNamesRemote()) do
        if is_storage(m, storage_name) then
            local size = m.callRemote(storage_name, 'size')

            local storage_items = {}
            local items = m.callRemote(storage_name, 'list')
            for slot in pairs(items) do
                local details = m.callRemote(storage_name, 'getItemDetail', slot)
                if details ~= nil then
                    local cache_item = { item = details, slot = slot }
                    table.insert(storage_items, cache_item)
                end
            end
            table.insert(storages, { name = storage_name, size = size, items = storage_items })
        end
    end
    return storages
end

return function(methods, handlers, wsclient)
    methods['getItems'] = getItems
    print(modem)
    print(7 * 2)
end
