package main

const (
	// Control Flow
	NodeEntry       = "Entry"
	NodeExit        = "Exit"
	NodeIf          = "If"
	NodeIfEnd       = "IfEnd"
	NodeIfElse      = "IfElse"
	NodeElseIf      = "ElseIf"
	NodeElseIfList  = "ElseIfList"
	NodeSwitch      = "Switch"
	NodeCase        = "Case"
	NodeDefault     = "Default"
	NodeWhile       = "While"
	NodeWhileEnd    = "WhileEnd"
	NodeDoWhile     = "DoWhile"
	NodeDoWhileEnd  = "DoWhileEnd"
	NodeFor         = "For"
	NodeForEach     = "ForEach"
	NodeForEnd      = "ForEnd"
	NodeForEachEnd  = "ForEachEnd"
	NodeTryCatch    = "TryCatch"
	NodeTryCatchEnd = "TryCatchEnd"
	NodeThrow       = "Throw"
	NodeBreak       = "Break"
	NodeContinue    = "Continue"

	// Function & Method Handling
	NodeFunctionCall    = "FunctionCall"
	NodeFunction        = "Function"
	NodeMethodCall      = "MethodCall"
	NodeMethod          = "Method"
	NodeReturn          = "Return"
	NodeReturnValue     = "ReturnValue"
	NodeReturnReference = "ReturnReference"
	NodeCallBegin       = "CallBegin"
	NodeCallEnd         = "CallEnd"
	NodeRetValue        = "RetValue"
	NodeId              = "id"
	NodeArgumentList    = "ArgumentList"
	NodeArgument        = "Argument"

	// Class & Interface Handling
	NodeClass         = "Class"
	NodeInterface     = "Interface"
	NodeImplements    = "Implements"
	NodeParentClass   = "ParentClass"
	NodeClassInstance = "ClassInstance"

	// Variables & Assignments
	NodeVariable            = "Variable"
	NodeAssign              = "Assign"
	NodeReferenceParam      = "ReferenceParam"
	NodeTypedValueParam     = "TypedValueParam"
	NodeTypedReferenceParam = "TypedReferenceParam"
	NodeValueParam          = "ValueParam"
	NodeParameterList       = "ParameterList"

	// Operators
	NodeBinOp         = "BinOP"
	NodeRelOp         = "RelOP"
	NodeLogicOp       = "LogicOP"
	NodeUnaryOp       = "UnaryOP"
	NodeIncrement     = "Increment"
	NodePostIncrement = "PostIncrement"
	NodePreIncrement  = "PreIncrement"

	// Expressions
	NodeCondition        = "Condition"
	NodeConditionalTrue  = "ConditionalTrue"
	NodeConditionalFalse = "ConditionalFalse"
	NodeTernary          = "Ternary"
	NodeTernaryEnd       = "TernaryEnd"
	NodeString           = "String"
	NodeStringExpr       = "StringExpression"
	NodeStringLiteral    = "StringLiteral"
	NodeInteger          = "Integer"
	NodeIntegerLiteral   = "IntegerLiteral"
	NodeDouble           = "Double"
	NodeHexLiteral       = "HexLiteral"
	NodeBool             = "Bool"
	NodeTrue             = "True"
	NodeFalse            = "False"
	NodeNull             = "Null"

	// Includes & Imports
	NodeInclude     = "Include"
	NodeIncludeOnce = "IncludeOnce"
	NodeRequire     = "Require"
	NodeRequireOnce = "RequireOnce"

	// Echo, Print & Output
	NodeEcho       = "Echo"
	NodePrint      = "Print"
	NodeExecString = "ExecString"
	NodeHeredoc    = "Heredoc"

	// Statements & Blocks
	NodeStatementBody  = "StatementBody"
	NodeBlock          = "Block"
	NodeExpressionStmt = "ExpressionStatement"
	NodeUnset          = "Unset"

	// Member Declarations
	NodeMemberDeclaration = "MemberDeclaration"
	NodePublicMember      = "PublicMember"
	NodePrivateMember     = "PrivateMember"
	NodeProtectedMember   = "ProtectedMember"
	NodeConstMember       = "ConstMember"
	NodeAbstractMethod    = "AbstractMethod"
	NodePublicMethod      = "PublicMethod"
	NodeProtectedMethod   = "ProtectedMethod"
	NodePrivateMethod     = "PrivateMethod"
	NodeReturnValueMethod = "ReturnValueMethod"

	// Miscellaneous
	NodeHtml      = "Html"
	NodeArobas    = "Arobas"
	NodeDead      = "Dead"
	NodeLegalChar = "LegalChar"
	NodeStart     = "Start"
)
