local input = peripheral.wrap 'left'--[[@as ccTweaked.peripherals.Inventory]]
local output = peripheral.wrap 'right'--[[@as ccTweaked.peripherals.Inventory]]
local orb_store = peripheral.wrap 'front'--[[@as ccTweaked.peripherals.Inventory]]
local altar = peripheral.wrap 'back'--[[@as ccTweaked.peripherals.Inventory]]
local term = peripheral.wrap 'top'--[[@as ccTweaked.peripherals.Monitor]]

local previous = nil

local function has_prefix(str, prefix)
    return string.sub(str, 1, string.len(prefix)) == prefix
end

local function has_suffix(str, suffix)
    return string.sub(str, -#suffix) == suffix
end

local function is_orb(item)
    if item == nil then
        return false
    end
    return has_prefix(item['name'], 'bloodmagic:') and has_suffix(item['name'], 'bloodorb')
end

local function is_changed(item)
    if previous == nil then
        return true
    end
    if previous['name'] == item['name'] then
        return false
    end

    print(previous['name'] .. ' != ' .. item['name'])
    return true
end

local function get_input_slot()
    for s in pairs(input.list()) do
        return s
    end
    return nil
end

local function tick()
    local blood_amount = 0
    for _, t in pairs(altar.tanks()) do
        blood_amount = t['amount']
    end
    term.setCursorPos(1, 1)
    term.clearLine()
    term.write(tostring(blood_amount))

    local current = altar.getItemDetail(1)
    local has_orb = is_orb(current)
    if current ~= nil and not has_orb then
        if not is_changed(current) then
            print 'still crafting'
            return true
        end
        print 'get result'
        output.pullItems('back', 1)
    end

    local input_slot = get_input_slot()

    if has_orb then
        if input_slot == nil then
            print 'keep orb'
            return true
        end
        orb_store.pullItems('back', 1)
        input.pushItems('back', input_slot, 1)
        previous = altar.getItemDetail(1)
        print 'replace orb by item'
        return true
    end

    if input_slot == nil then
        orb_store.pushItems('back', 1)
        print 'try charge orb'
    else
        input.pushItems('back', input_slot, 1)
        previous = altar.getItemDetail(1)
        print 'charge item'
    end
end

return function(methods, handlers, wsclient)
    local timerDuration = 1
    local timerId = os.startTimer(timerDuration)

    handlers['timer'] = function(eventData)
        if eventData[2] == timerId then
            tick()
            timerId = os.startTimer(timerDuration)
        end
    end
    return {}
end
