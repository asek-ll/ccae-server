local glass = {
    name = 'minecraft:glass',
    count = 1,
    maxCount = 64,
}
testnet.registerRecipe {
    result = {
        name = 'minecraft:glass_pane',
        count = 16,
        maxCount = 64,
    },
    ingredients = {
        [1] = glass,
        [2] = glass,
        [3] = glass,
        [4] = glass,
        [5] = glass,
        [6] = glass,
    },
}
return function(network, crafter)
    testnet.enableTurtleApi(crafter)
    local crafter_modem = testnet.createModem(network)
    local crafter_input = testnet.createInventory(network, 'crafter_input')
    local crafter_output = testnet.createInventory(network, 'crafter_output')
    local crafter_buffer = testnet.createInventory(network, 'crafter_buffer')
    -- testnet.setItem(crafter_buffer, 1, {
    --     name = 'resourcefulbees:bee_jar',
    --     count = 1,
    --     maxCount = 64,
    -- })
    ccemux.openEmu(crafter)
    testnet.attachPeripheral('top', crafter_input, crafter)
    testnet.attachPeripheral('bottom', crafter_buffer, crafter)
    testnet.attachPeripheral('front', crafter_output, crafter)
    testnet.attachPeripheral('left', crafter_modem, crafter)
    testnet.attachWorkbench('right', crafter)

    -- print(crafter_buffer)
end
