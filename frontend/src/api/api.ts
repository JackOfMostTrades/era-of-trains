import {TrackTile} from "../game/map_state.ts";

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
    minPlayers: number;
    maxPlayers: number;
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
}

export function WhoAmI(req: WhoAmIRequest): Promise<WhoAmIResponse> {
    return doApiCall('/api/whoami', req);
}

export interface LoginRequest {
    provider?: string
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
    provider: string
    accessToken: string
    nickname: string
}
export interface RegisterResponse {
}

export function Register(req: RegisterRequest): Promise<RegisterResponse> {
    return doApiCall('/api/register', req);
}

export interface LinkProfileRequest {
    provider: string
    accessToken: string
}
export interface LinkProfileResponse {
}
export function LinkProfile(req: LinkProfileRequest): Promise<LinkProfileResponse> {
    return doApiCall('/api/linkProfile', req);
}

export interface LogoutRequest {}
export interface LogoutResponse {}

export function Logout(req: LogoutRequest): Promise<LogoutResponse> {
    return doApiCall('/api/logout', req);
}

export interface CreateGameRequest {
    name: string;
    minPlayers: number;
    maxPlayers: number;
    mapName: string;
    inviteOnly: boolean;
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
    WHITE,
}
export const Colors = [Color.BLACK, Color.RED, Color.YELLOW, Color.BLUE, Color.PURPLE, Color.WHITE];

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
export const PlayerColors = [PlayerColor.BLUE, PlayerColor.GREEN, PlayerColor.YELLOW, PlayerColor.PINK, PlayerColor.GRAY, PlayerColor.ORANGE];

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
    urbanizations: Urbanization[]|undefined;

    // Map from color to number of cubes of that color in the bag
    cubeBag: { [ color: string ]: number }
    // Cubes present on the board
    cubes?: BoardCube[];
    // Cubes present on the goods-growth chart, 1-6 white, 7-12 black, 13-20 new cities
    goodsGrowth: Color[][]
    // If cubes have been drawn for the production action, these are the cubes
    productionCubes: Color[]

    // State specific to the map
    mapState: any|undefined
}
export interface ViewGameRequest {
    gameId: string;
}
export interface ViewGameResponse {
    id: string;
    name: string;
    started: boolean;
    finished: boolean;
    minPlayers: number;
    maxPlayers: number;
    mapName: string;
    ownerUser: User;
    activePlayer: string;
    joinedUsers: User[];
    gameState?: GameState;
    inviteOnly: boolean;
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

export interface DeleteGameRequest {
    gameId: string
}
export interface DeleteGameResponse {
}
export function DeleteGame(req: DeleteGameRequest): Promise<DeleteGameResponse> {
    return doApiCall('/api/deleteGame', req);
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
    track: Direction[];
}

export interface TrackPlacement {
    tile: TrackTile;
    rotation: number;
}

export interface TeleportLinkPlacement {
    track: Direction;
}

export interface BuildStep {
    hex: Coordinate;

    // One of...
    urbanization?: number|undefined;
    townPlacement?: TownPlacement;
    trackPlacement?: TrackPlacement;
    teleportLinkPlacement?: TeleportLinkPlacement;
}

export interface BuildAction {
    steps?: BuildStep[]
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
    reversible: boolean;
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

export interface CustomColors {
    playerColors?: Array<string|undefined>
    goodsColors?: Array<string|undefined>
}

export interface GetMyProfileRequest {
}
export interface GetMyProfileResponse {
    id: string;
    nickname: string;
    email: string;
    googleId?: string;
    discordId?: string;
    emailNotificationsEnabled: boolean;
    discordTurnAlertsEnabled: boolean;
    colorPreferences?: PlayerColor[];
    customColors?: CustomColors;
    webhooks: string[];
}
export function GetMyProfile(req: GetMyProfileRequest): Promise<GetMyProfileResponse> {
    return doApiCall('/api/getMyProfile', req);
}

export interface SetMyProfileRequest {
    emailNotificationsEnabled?: boolean;
    discordTurnAlertsEnabled?: boolean;
    colorPreferences?: PlayerColor[];
    customColors?: CustomColors;
    webhooks?: string[];
}
export interface SetMyProfileResponse {
}
export function SetMyProfile(req: SetMyProfileRequest): Promise<SetMyProfileResponse> {
    return doApiCall('/api/setMyProfile', req);
}

export interface GetGameChatRequest {
    gameId: string;
    after?: number;
}
export interface GameChatMessage {
    userId: string;
    timestamp: number;
    message: string;
}
export interface GetGameChatResponse {
    messages: GameChatMessage[]|undefined;
}
export function GetGameChat(req: GetGameChatRequest): Promise<GetGameChatResponse> {
    return doApiCall('/api/getGameChat', req);
}

export interface SendGameChatRequest {
    gameId: string;
    message: string;
}
export interface SendGameChatResponse {
}
export function SendGameChat(req: SendGameChatRequest): Promise<SendGameChatResponse> {
    return doApiCall('/api/sendGameChat', req);
}

export interface PollGameStatusRequest {
    gameId: string;
}
export interface PollGameStatusResponse {
    lastMove: number;
    lastChat: number;
}
export function PollGameStatus(req: PollGameStatusRequest): Promise<PollGameStatusResponse> {
    return doApiCall('/api/pollGameStatus', req);
}

export interface UndoMoveRequest {
    gameId: string;
}
export interface UndoMoveResponse {
}
export function UndoMove(req: UndoMoveRequest): Promise<UndoMoveResponse> {
    return doApiCall('/api/undoMove', req);
}
