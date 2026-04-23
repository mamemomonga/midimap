
local L6Channel = 0 -- L6Max MIDI Ch.1

-- フェーダー -> ミキサーレベル
local remap_l6chs = {
    --  [入力CC] = 出力CC,

    -- フェーダ -> レベル
    [1] = {81}, -- L6max Mix: ch1
    [2] = {82}, -- L6max Mix: ch2
    [3] = {83}, -- L6max Mix: ch3
    [4] = {84}, -- L6max Mix: ch4
    [5] = {85}, -- L6max Mix: ch5
    [6] = {86}, -- L6max Mix: ch6
    [7] = {87}, -- L6max Mix: ch7
    [8] = {88}, -- L6max Mix: ch8

    -- ミュートボタン(0:Unmute / 127:Mute)
    [21] = {93 }, -- L6max Mute: ch1
    [22] = {94 }, -- L6max Mute: ch2
    [23] = {95 }, -- L6max Mute: ch3
    [24] = {102}, -- L6max Mute: ch4
    [25] = {103}, -- L6max Mute: ch5
    [26] = {104}, -- L6max Mute: ch6
    [27] = {105}, -- L6max Mute: ch7
    [28] = {106}, -- L6max Mute: ch8

    -- ノブ -> EFXレベル
    [13] = {65}, -- L6max EFX: ch1
    [14] = {66}, -- L6max EFX: ch2
    [15] = {67}, -- L6max EFX: ch3
    [16] = {68}, -- L6max EFX: ch4
    [17] = {69}, -- L6max EFX: ch5
    [18] = {70}, -- L6max EFX: ch6
    [19] = {71}, -- L6max EFX: ch7
    [20] = {72}, -- L6max EFX: ch8
}

function on_note_on(ch, note, vel)
end

function on_note_off(ch, note, vel)
end

function on_cc(ch, cc, val)
    local l6chs = remap_l6chs[cc]
    if l6chs then
        for _, dcc in ipairs(l6chs) do
            send_cc(0, dcc, val)
        end
    end
end
