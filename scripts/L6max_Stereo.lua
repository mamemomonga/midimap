
local L6Channel = 0 -- L6Max MIDI Ch.1

-- フェーダー -> ミキサーレベル
local remap_l6chs = {
    --  [入力CC] = 出力CC,

    -- フェーダ -> レベル
    [1] = {81, 82}, -- L6max Mix: ch1, ch2
    [2] = {83, 84}, -- L6max Mix: ch3, ch4
    [3] = {85},     -- L6max Mix: ch5
    [4] = {86},     -- L6max Mix: ch6
    [5] = {87},     -- L6max Mix: ch7
    [6] = {88},     -- L6max Mix: ch8

    -- ミュートボタン(0:Unmute / 127:Mute)
    [21] = {93, 94 }, -- L6max Mute: ch1, ch2
    [22] = {95, 102}, -- L6max Mute: ch3, ch4
    [23] = {103},     -- L6max Mute: ch5
    [24] = {104},     -- L6max Mute: ch6
    [25] = {105},     -- L6max Mute: ch7
    [26] = {106},     -- L6max Mute: ch8

    -- ノブ -> EFXレベル
    [13] = {65, 66}, -- L6max EFX: ch1, ch2
    [14] = {67, 68}, -- L6max EFX: ch3, ch4
    [15] = {69},     -- L6max EFX: ch5
    [16] = {70},     -- L6max EFX: ch6
    [17] = {71},     -- L6max EFX: ch7
    [18] = {72},     -- L6max EFX: ch8
}

function on_note_on(ch, note, vel)
end

function on_note_off(ch, note, vel)
end

function on_cc(ch, cc, val)
    local l6chs = remap_l6chs[cc]
    if l6chs then
        for _, dcc in ipairs(l6chs) do
            send_cc(L6Channel, dcc, val)
        end
    end
end
