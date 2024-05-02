periphemu.create("top", "modem")

local mod = peripheral.wrap("top")
print(mod)

exit(1)

local loginMessage = textutils.serializeJSON {
    jsonrpc = '2.0',
    id = 1,
    method = 'login',
    params = {
        id = 'randomUniqueId101',
        role = 'storage',
    },
}

local methods = {
    eval = function(body)
        local f, msg = load(body)
        if f == nil then
            print(msg)
        end
        return f()
    end,
}

local function listen()
    local url = 'ws://localhost:12526'
    local ws = assert(http.websocket(url))
    ws.send(loginMessage)

    while true do
        local eventData = { os.pullEvent() }
        local event = eventData[1]

        if event == 'websocket_message' then
            print('Received message from ' .. eventData[2] .. ' with contents ' .. eventData[3])
            local data = textutils.unserializeJSON(eventData[3])
            local method = data['method']
            local params = data['params']
            if methods[method] ~= nil then
                local state, result = pcall(methods[method], params)
                local message
                if state then
                    message = textutils.serializeJSON {
                        jsonrpc = '2.0',
                        id = data['id'],
                        result = result,
                    }
                else
                    message = textutils.serializeJSON {
                        jsonrpc = '2.0',
                        id = data['id'],
                        error = {
                            code = -1,
                            message = result,
                        },
                    }
                end
                print('Send message', message)
                ws.send(message)
            end
        end

        if event == 'websocket_failure' then
            print('Websocket failure from ' .. eventData[2] .. ' with contents ' .. eventData[3])
            ws.close()
            error(eventData[3])
        end

        if event == 'websocket_closed' then
            print('Websocket closed from ' .. eventData[2])
            ws.close()
            error(eventData[3])
        end
    end
end

while true do
    status, err = pcall(listen)
    if not status then
        print('Got error' .. err)
        os.sleep(10)
    end
end
