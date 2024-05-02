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
        local f = load(body)
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
            local data = textutils.unserilizeJSON(eventData[3])
            local method = data['method']
            local params = data['params']
            if methods[method] ~= nil then
                state, result = pcal(methods[method], params)
                if state then
                    ws.send(textutils.serializeJSON {
                        jsonrpc = '2.0',
                        id = data['id'],
                        error = {
                            code = -1,
                            message = result,
                        },
                    })
                else
                    ws.send(textutils.serializeJSON {
                        jsonrpc = '2.0',
                        id = data['id'],
                        result = result,
                    })
                end
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
