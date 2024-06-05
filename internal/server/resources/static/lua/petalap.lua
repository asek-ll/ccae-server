local input = peripheral.wrap 'right' --[[@as ccTweaked.peripherals.Inventory]]
local output = peripheral.wrap 'left' --[[@as ccTweaked.peripherals.Inventory]]
local buf = 'top'

local function has_input()
    for _, item in pairs(input.list()) do
        if item ~= nil then
            return true
        end
    end
    return false
end

local function craft()
    if not has_input() then
        return false
    end

    turtle.select(1)
    local b = turtle.getItemDetail()
    if b == nil then
        return false
    end
    if b['name'] ~= 'minecraft:water_bucket' then
        turtle.placeDown()
    end

    turtle.drop()
    turtle.suck()

    turtle.select(2)

    for slot, item in pairs(input.list()) do
        if item ~= nil then
            input.pushItems(buf, slot, 1)
            turtle.suckUp()
            turtle.drop()
        end
    end
    turtle.suck()
    turtle.dropUp()

    output.pullItems(buf, 1, 64)
    return true
end

---@param handlers table<string,fun(eventData:table)>
return function(_, handlers, _)
    local timerDuration = 30
    local timerId = os.startTimer(timerDuration)

    handlers['timer'] = function(eventData)
        print 'ON TIMER'
        if eventData[2] == timerId then
            while craft() do
            end
            timerId = os.startTimer(timerDuration)
        end
    end

    return {}
end
