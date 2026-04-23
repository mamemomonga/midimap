-- TransformCC: 入力CCが指定したCCと比較CCが合致したら、出力CCで入力Valを出力する
-- 比較CC: MIDIコントローラのCC
-- 出力CC: L6maxのCC
--   local chc=TransformCC.new(入力CC, 入力Val, 出力Ch)
--   chc(比較CC, 出力CC)
local TransformCC = {}
function TransformCC.new(cc, val, nch)
    return function(mcc, ncc)
        if(cc == mcc) then
            send_cc(nch, ncc, val)
        end
    end
end

function on_note_on(ch, note, vel)
end

function on_note_off(ch, note, vel)
end

function on_cc(ch, cc, val)
    local chc = TransformCC.new(cc, val, 1)

    -- フェーダー -> レベル
    chc(  1,  81); -- ch1
    chc(  1,  82); -- ch2
    chc(  2,  83); -- ch3
    chc(  2,  84); -- ch4
    chc(  3,  85); -- ch5
    chc(  4,  86); -- ch6
    chc(  5,  87); -- ch7
    chc(  6,  88); -- ch8

    -- ミュート(0:Unmute / 127:Mute)
    chc(  21, 93); -- ch1
    chc(  21, 94); -- ch2
    chc(  22, 95); -- ch3
    chc(  22,102); -- ch4
    chc(  23,103); -- ch5
    chc(  24,104); -- ch6
    chc(  25,105); -- ch7
    chc(  26,106); -- ch8

    -- ノブ(EFX)
    chc(  13, 65); -- ch1
    chc(  13, 66); -- ch2
    chc(  14, 67); -- ch3
    chc(  14, 68); -- ch4
    chc(  15, 69); -- ch5
    chc(  16, 70); -- ch6
    chc(  17, 71); -- ch7
    chc(  18, 72); -- ch8

end
