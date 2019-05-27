pragma solidity ^0.4.0;

contract Check {

    struct check {
        uint Id;
        uint Amount;
    }

    event InputLog (
        uint Id,
        uint Amount
        );

    mapping (uint => check) checkInfos;

    function createCheckInfos( uint Id, uint Amount) {
        InputLog(Id,Amount);
        check c;
        c.Id = Id;
        c.Amount = Amount;

        checkInfos[Id] = c;
    }

    function getPremiumInfos(uint Id) public constant returns(string,uint) {
        return (
        checkInfos[Id].Amount
        );
    }
}