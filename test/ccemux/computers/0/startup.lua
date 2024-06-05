local m1 = testnet.createModem(1)

local i1 = testnet.createInventory(1)
local i2 = testnet.createInventory(1)

testnet.setItem(i1, 1, {
    name = 'minecraft:apple',
    count = 2,
    maxCount = 64,
})

testnet.setItem(i2, 2, {
    name = 'minecraft:ender_pearl',
    count = 3,
    maxCount = 16,
})

testnet.setItem(i2, 4, {
    name = 'pneumaticcraft:advanced_pcb',
    count = 30,
    maxCount = 64,
})

testnet.setItem(i2, 5, {
    name = 'pneumaticcraft:printed_circuit_board',
    count = 60,
    maxCount = 64,
})

testnet.setItem(i2, 6, {
    name = 'minecraft:glass',
    count = 64,
    maxCount = 64,
})

ccemux.openEmu(1)
testnet.attachPeripheral('top', m1, 1)

local crafter = 2
local crafter_modem = testnet.createModem(1)
local crafter_buffer = testnet.createInventory(1)
testnet.setItem(crafter_buffer, 1, {
    name = 'resourcefulbees:bee_jar',
    count = 1,
    maxCount = 64,
})
ccemux.openEmu(crafter)
testnet.attachPeripheral('top', crafter_buffer, crafter)
testnet.attachPeripheral('left', crafter_modem, crafter)

ccemux.closeEmu()
