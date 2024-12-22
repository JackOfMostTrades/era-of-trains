
async function doApiCall<ReqT, ResT>(requestPath: string, req: ReqT): Promise<ResT> {
    let res = await fetch(requestPath, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(req)
    })
    if (!res.ok) {
        let body = await res.text();
        throw new Error("got non-ok response: " + body);
    }
    return await res.json();
}

export interface ListGamesRequest {
}
export interface GameSummary {
    id: string;
    name: string;
    started: boolean;
    finished: boolean;
    numPlayers: number;
    mapName: string;
    activePlayer: string;
    ownerUser: User;
    joinedUsers: User[];
}
export interface ListGamesResponse {
    games?: GameSummary[];
}

export function ListGames(req: ListGamesRequest): Promise<ListGamesResponse> {
    return doApiCall('/api/listGames', req);
}

export interface GetMyGamesRequest {}
export interface GetMyGamesResponse {
    games?: GameSummary[];
}

export function GetMyGames(req: GetMyGamesRequest): Promise<GetMyGamesResponse> {
    return doApiCall('/api/getMyGames', req)
}

export interface WhoAmIRequest {

}
export interface WhoAmIResponse {
    user: User
    waitingForMeCount: number
}

export function WhoAmI(req: WhoAmIRequest): Promise<WhoAmIResponse> {
    return doApiCall('/api/whoami', req);
}

export interface LoginRequest {
    accessToken?: string
    devNickname?: string
}
export interface LoginResponse {
    registrationRequired: boolean
}

export function Login(req: LoginRequest): Promise<LoginResponse> {
    return doApiCall('/api/login', req);
}

export interface RegisterRequest {
    accessToken: string
    nickname: string
}
export interface RegisterResponse {
}

export function Register(req: RegisterRequest): Promise<RegisterResponse> {
    return doApiCall('/api/register', req);
}

export interface LogoutRequest {}
export interface LogoutResponse {}

export function Logout(req: LogoutRequest): Promise<LogoutResponse> {
    return doApiCall('/api/logout', req);
}

export interface CreateGameRequest {
    name: string;
    numPlayers: number;
    mapName: string;
}
export interface CreateGameResponse {
    id: string;
}

export function CreateGame(req: CreateGameRequest): Promise<CreateGameResponse> {
    return doApiCall('/api/createGame', req);
}

export type SpecialAction = 'first_move' | 'first_build' | 'engineer' | 'loco' | 'urbanization' | 'production' | 'turn_order_pass'

export interface User {
    nickname: string;
    email: string;
    id: string;
}
export interface Coordinate {
    x: number
    y: number
}
export enum GamePhase {
    SHARES  = 1,
    AUCTION ,
    CHOOSE_SPECIAL_ACTIONS ,
    BUILDING ,
    MOVING_GOODS ,
    GOODS_GROWTH ,
}
export enum Direction {
    NORTH = 0,
    NORTH_EAST,
    SOUTH_EAST,
    SOUTH,
    SOUTH_WEST,
    NORTH_WEST
}
export const ALL_DIRECTIONS = [Direction.NORTH, Direction.NORTH_EAST, Direction.SOUTH_EAST, Direction.SOUTH, Direction.SOUTH_WEST, Direction.NORTH_WEST];
export enum Color {
    NONE = 0,
    BLACK,
    RED,
    YELLOW,
    BLUE,
    PURPLE,
}
export interface Link {
    sourceHex: Coordinate;
    steps:     Direction[];
    owner: string;
    complete: boolean;
}
export interface Urbanization {
    hex: Coordinate;
    city: number;
}
export interface BoardCube {
    color: Color;
    hex: Coordinate;
}

export enum PlayerColor {
    BLUE = 0,
    GREEN,
    YELLOW,
    PINK,
    GRAY,
    ORANGE
}

export interface GameState {
    playerOrder: string[];
    playerColor: { [playerId: string]: PlayerColor }
    playerShares: { [playerId: string]: number }
    playerLoco: { [playerId: string]: number }
    playerIncome: { [playerId: string]: number }
    playerActions:  { [playerId: string]: SpecialAction }
    playerCash: { [playerId: string]: number }

    // Map from player ID to their last bid
    auctionState?: { [playerId: string]: number }

    gamePhase: GamePhase;
    turnNumber: number;

    // Which round of moving goods are we in (0 or 1)
    movingGoodsRound: number;
    // Which users did loco during move goods (to ensure they don't double-loco)
    playerHasDoneLoco: { [playerId: string]: boolean }
    links: Link[];
    urbanizations: Urbanization[];

    // Map from color to number of cubes of that color in the bag
    cubeBag: { [ color: string ]: number }
    // Cubes present on the board
    cubes?: BoardCube[];
    // Cubes present on the goods-growth chart, 1-6 white, 7-12 black, 13-20 new cities
    goodsGrowth: Color[][]
    // If cubes have been drawn for the production action, these are the cubes
    productionCubes: Color[]
}
export interface ViewGameRequest {
    gameId: string;
}
export interface ViewGameResponse {
    id: string;
    name: string;
    started: boolean;
    finished: boolean;
    numPlayers: number;
    mapName: string;
    ownerUser: User;
    activePlayer: string;
    joinedUsers: User[];
    gameState?: GameState;
}

export function ViewGame(req: ViewGameRequest): Promise<ViewGameResponse> {
    return doApiCall('/api/viewGame', req);
}

export interface JoinGameRequest {
    gameId: string
}
export interface JoinGameResponse {
}
export function JoinGame(req: JoinGameRequest): Promise<JoinGameResponse> {
    return doApiCall('/api/joinGame', req);
}

export interface LeaveGameRequest {
    gameId: string
}
export interface LeaveGameResponse {
}
export function LeaveGame(req: LeaveGameRequest): Promise<LeaveGameResponse> {
    return doApiCall('/api/leaveGame', req);
}

export interface StartGameRequest {
    gameId: string
}
export interface StartGameResponse {
}
export function StartGame(req: StartGameRequest): Promise<StartGameResponse> {
    return doApiCall('/api/startGame', req);
}

type ActionName = 'shares' | 'bid' | 'choose_action' | 'build' | 'move_goods' | 'produce_goods';

export interface SharesAction {
    amount: number;
}

export interface BidAction {
    amount: number
}

export interface ChooseAction {
    action: SpecialAction;
}

export interface TownPlacement {
    track: Direction;
    hex: Coordinate;
}

export interface TrackPlacement {
    track: [Direction, Direction];
    hex: Coordinate;
}

export interface BuildAction {
    townPlacements: TownPlacement[];
    trackPlacements: TrackPlacement[];
    urbanization?: Urbanization;
}

export interface MoveGoodsAction {
    startingLocation?: Coordinate;
    color?: Color;
    path?: Direction[];
    loco?: boolean;
}

export interface ProduceGoodsAction {
    // List (corresponding the cubes in the same order as ProductionCubes in the game state) with X,Y coordinates
    // corresponding to which city (X) and which spot (Y) within that city
    destinations: Coordinate[];
}

export interface ConfirmMoveRequest {
    gameId: string
    actionName: ActionName
    sharesAction?: SharesAction
    bidAction?: BidAction
    chooseAction?: ChooseAction
    buildAction?: BuildAction
    moveGoodsAction?: MoveGoodsAction
    produceGoodsAction?: ProduceGoodsAction
}
export interface ConfirmMoveResponse {
}

export function ConfirmMove(req: ConfirmMoveRequest): Promise<ConfirmMoveResponse> {
    return doApiCall('/api/confirmMove', req);
}


export interface GameLogEntry {
    timestamp: number;
    userId: string;
    action: string;
    description: string;
}

export interface GetGameLogsRequest {
    gameId: string;
}
export interface GetGameLogsResponse {
    logs?: GameLogEntry[];
}
export function GetGameLogs(req: GetGameLogsRequest): Promise<GetGameLogsResponse> {
    return doApiCall('/api/getGameLogs', req);
}
