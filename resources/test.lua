message = textutils.serializeJSON {
    jsonrpc = '2.0',
    id = 1,
    method = 'login',
    params = {
        id = 'randomUniqueId101',
        role = 'storage',
    },
}

function listen()
    local ws = assert(http.websocket 'ws://localhost:12526')
    ws.send(message)

    while true do
        local eventData = { os.pullEvent() }
        local event = eventData[1]

        if event == 'websocket_message' then
            print('Received message from ' .. eventData[2] .. ' with contents ' .. eventData[3])
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
