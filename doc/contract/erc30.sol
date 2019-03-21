pragma solidity ^0.4.21;

/**
 * @title SafeMath
 * @dev Math operations with safety checks that throw on error
 */
library SafeMath {

  /**
  * @dev Multiplies two numbers, throws on overflow.
  */
  function mul(uint256 a, uint256 b) internal pure returns (uint256 c) {
    if (a == 0) {
      return 0;
    }
    c = a * b;
    assert(c / a == b);
    return c;
  }

  /**
  * @dev Integer division of two numbers, truncating the quotient.
  */
  function div(uint256 a, uint256 b) internal pure returns (uint256) {
    // assert(b > 0); // Solidity automatically throws when dividing by 0
    // uint256 c = a / b;
    // assert(a == b * c + a % b); // There is no case in which this doesn't hold
    return a / b;
  }

  /**
  * @dev Subtracts two numbers, throws on overflow (i.e. if subtrahend is greater than minuend).
  */
  function sub(uint256 a, uint256 b) internal pure returns (uint256) {
    assert(b <= a);
    return a - b;
  }

  /**
  * @dev Adds two numbers, throws on overflow.
  */
  function add(uint256 a, uint256 b) internal pure returns (uint256 c) {
    c = a + b;
    assert(c >= a);
    return c;
  }
}

contract StandardToken {
  using SafeMath for uint256;

  string public name;                 
  uint8 public decimals;               
  string public symbol;              
  address public owner;
  
  mapping(address => uint256) balances;
  mapping(address => uint256) nounce;
  mapping(address => string [] ) coins;
  address [] ensurances;
  
  struct Tx {
    address from;
    address to;
	uint256 offset;
    uint256 value;     	
  }
  mapping(string => Tx ) txs;
  

  uint256 totalSupply_;

  mapping (address => mapping (address => uint256)) internal allowed;
  
  constructor(uint256 _initialAmount, string _tokenName, uint8 _decimalUnits, string _tokenSymbol) public {
    totalSupply_ = _initialAmount * 10 ** uint256(_decimalUnits);
    balances[msg.sender] = totalSupply_; 

	owner = msg.sender;
    name = _tokenName;                   
    decimals = _decimalUnits;          
    symbol = _tokenSymbol;
  }
  
  /**
  * @dev total number of tokens in existence
  */
  function totalSupply() public view returns (uint256) {
    return totalSupply_;
  }

  /**
  * @dev transfer token for a specified address
  * @param _to The address to transfer to.
  * @param _value The amount to be transferred.
  */
  function transfer(address _to, uint256 _value) public returns (bool) {
    require(_to != address(0));
	uint256 balance;
	uint256 withdraw;
	(balance, withdraw) = mergeBalanceOf(msg.sender);
    require(_value <= balance);

    balances[msg.sender] = balances[msg.sender].sub(_value);
    balances[_to] = balances[_to].add(_value);
    emit Transfer(msg.sender, _to, _value);
    return true;
  }
  
  function transfer_ensure(address _to, uint256 _value) public returns (bool) {
    require(_to != address(0));
	uint256 balance;
	uint256 withdraw;
	(balance, withdraw) = mergeBalanceOf(msg.sender);
    require(_value <= balance);
	if (balances[_to] == 0){
	    ensurances.push(_to);
	}
    balances[msg.sender] = balances[msg.sender].sub(_value);
    balances[_to] = balances[_to].add(_value);
	
    emit Transfer(msg.sender, _to, _value);
    return true;
  }

  /**
  * @dev Gets the balance of the specified address.
  * @param _owner The address to query the the balance of.
  * @return An uint256 representing the amount owned by the passed address.
  */
  function balanceOf(address _owner) public view returns (uint256, uint256) {
    uint256 balance = balances[_owner];
	uint256 delay = 0;
    uint len = coins[_owner].length; 
    	
    for (uint n = 0; n < len; n++) {
        if ( keccak256(coins[_owner][n]) == keccak256(""))
          continue;
          
        if (txs[coins[_owner][n]].offset > block.number ){
			
          delay = delay + txs[coins[_owner][n]].value;
		}
		else if (txs[coins[_owner][n]].to == _owner){
		       balance = balance + txs[coins[_owner][n]].value;
		}
    }
	
	return (balance, delay);
  }

  
  function mergeBalanceOf(address _owner) returns (uint256, uint256) {
    uint256 balance = balances[_owner];
	uint256 delay = 0;
    uint len = coins[_owner].length; 
    	
	bool [] memory deleted = new bool[](len);
    for (uint n = 0; n < len; n++) {
        if ( keccak256(coins[_owner][n]) == keccak256(""))
          continue;
        
		if ( txs[coins[_owner][n]].offset == 0)
          continue;			
		  
        if (txs[coins[_owner][n]].offset > block.number ){
			
          delay = delay + txs[coins[_owner][n]].value;
		  deleted[n] = false;
		}
		else{
			deleted[n] = true;
			if (txs[coins[_owner][n]].to == _owner){
		       balance = balance + txs[coins[_owner][n]].value;
		       balances[_owner] = balance;
			}
		}
    }
	
	for (n = 0; n < len; n++){
	  if (deleted[n])	
	    delete(coins[_owner][n]);
	}
	
	return (balance, delay);
  }


  /**
   * @dev Transfer tokens from one address to another
   * @param _from address The address which you want to send tokens from
   * @param _to address The address which you want to transfer to
   * @param _value uint256 the amount of tokens to be transferred
   */
  function transferFrom(address _from, address _to, uint256 _value) public returns (bool) {
    require(_to != address(0));
	uint256 balance;
	uint256 withdraw;
	(balance, withdraw) = mergeBalanceOf(_from);
    require(_value <= balance);
    require(_value <= allowed[_from][msg.sender]);

    balances[_from] = balances[_from].sub(_value);
    balances[_to] = balances[_to].add(_value);
    allowed[_from][msg.sender] = allowed[_from][msg.sender].sub(_value);
    emit Transfer(_from, _to, _value);
    return true;
  }

  /**
   * @dev Approve the passed address to spend the specified amount of tokens on behalf of msg.sender.
   *
   * Beware that changing an allowance with this method brings the risk that someone may use both the old
   * and the new allowance by unfortunate transaction ordering. One possible solution to mitigate this
   * race condition is to first reduce the spender's allowance to 0 and set the desired value afterwards:
   * https://github.com/ethereum/EIPs/issues/20#issuecomment-263524729
   * @param _spender The address which will spend the funds.
   * @param _value The amount of tokens to be spent.
   */
  function approve(address _spender, uint256 _value) public returns (bool) {
    allowed[msg.sender][_spender] = _value;
    emit Approval(msg.sender, _spender, _value);
    return true;
  }

  /**
   * @dev Function to check the amount of tokens that an owner allowed to a spender.
   * @param _owner address The address which owns the funds.
   * @param _spender address The address which will spend the funds.
   * @return A uint256 specifying the amount of tokens still available for the spender.
   */
  function allowance(address _owner, address _spender) public view returns (uint256) {
    return allowed[_owner][_spender];
  }
  
  event Approval(address indexed owner, address indexed spender, uint256 value);
  
  event Transfer(address indexed from, address indexed to, uint256 value);
  
  
  function transfer_withdrawable(address _to, uint256 _value) public returns (bool) {
    require(_to != address(0));
	uint256 balance;
	uint256 withdraw;
	(balance, withdraw) = mergeBalanceOf(msg.sender);
    require(_value <= balance);

	// TODO: check sender ensurance
	nounce[msg.sender] = nounce[msg.sender] + 1;
	
	string memory txid = stringAdd(addressToString(msg.sender), uintToString(nounce[msg.sender]));
	
    txs[txid] = Tx({
		from : msg.sender,
		to : _to,
		value : _value,
		offset : block.number + 200
	});
	
	
    balances[msg.sender] = balances[msg.sender].sub(_value);
	string  [] storage sender = coins[msg.sender];
	string [] storage reciever = coins[_to];
	sender.push(txid);
	reciever.push(txid);
	coins[msg.sender] = sender;
	coins[_to] = reciever;
	
    emit Transfer(msg.sender, _to, _value);
    return true;
  }
  
  function clear() public returns(bool) {
    require(owner == msg.sender);
	
	uint256 balance;
	uint256 withdraw;
    for (uint c = 0; c < ensurances.length; c++) {
		(balance, withdraw) = mergeBalanceOf(ensurances[c]);
		if (balance <= 0)
		    continue;
		// get owner address
		balances[ensurances[c]] = balances[ensurances[c]].sub(balance);
        balances[owner] = balances[owner].add(balance);
        emit Transfer(ensurances[c], owner, balance);
	}
  }
  
  function withdrawTx(string txid) public returns (bool) {
    Tx memory tx = txs[txid];
	require(0 != tx.value);
	require(msg.sender == tx.from);
	
	balances[tx.from] = balances[tx.from].add(tx.value);
	
	delete(txs[txid]);
	for (uint index = 0; index < coins[tx.from].length; index++) {
		if (keccak256(txid) == keccak256(coins[tx.from][index])){
		    delete(coins[tx.from][index]);
			break;
		}
    }
	
    for (uint i = 0; i < coins[tx.to].length; i++) {
		if (keccak256(txid) == keccak256(coins[tx.to][i])){
		    delete(coins[tx.to][i]);
			break;
		}
    }
	// get safety account 
    emit Transfer(tx.to, tx.from, tx.value);
    return true;
  }
  
  function getWithdrawableTX(address _owner) public view returns (string){
    uint len = coins[_owner].length; 
	if (len == 0)
	    return "";
	  
	require(coins[_owner].length > 0);
	
	string memory txids = "";
    for (uint n = 0; n < len; n++) {
        if ( keccak256(coins[_owner][n]) == keccak256(""))
          continue;
          
        if (txs[coins[_owner][n]].offset < block.number )
          continue;
          
		txids = coins[_owner][n];
		break;
    }
	
    return txids;
  }
  
  function stringAdd(string a, string b) returns(string){
    bytes memory _a = bytes(a);
    bytes memory _b = bytes(b);
    bytes memory res = new bytes(_a.length + _b.length);
    for(uint i = 0;i < _a.length;i++)
        res[i] = _a[i];
    for(uint j = 0;j < _b.length;j++)
        res[_a.length+j] = _b[j];  
    return string(res);
  }
  
  function uintToString(uint i) internal pure returns (string){
    if (i == 0) return "0";
    uint j = i;
    uint length;
    while (j != 0){
        length++;
        j /= 10;
    }
    bytes memory bstr = new bytes(length);
    uint k = length - 1;
    while (i != 0){
        bstr[k--] = byte(48 + i % 10);
        i /= 10;
    }
    return string(bstr);
  }
  
function addressToString(address _addr) public pure returns(string) {
    bytes32 value = bytes32(uint256(_addr));
    bytes memory alphabet = "0123456789abcdef";

    bytes memory str = new bytes(42);
    str[0] = '0';
    str[1] = 'x';
    for (uint i = 0; i < 20; i++) {
        str[2+i*2] = alphabet[uint(value[i + 12] >> 4)];
        str[3+i*2] = alphabet[uint(value[i + 12] & 0x0f)];
    }
    return string(str);
  }  
}