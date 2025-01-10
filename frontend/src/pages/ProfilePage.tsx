import {
    Button,
    Checkbox,
    Form,
    FormButton,
    FormField,
    FormInput,
    Header,
    Icon,
    Input,
    Label,
    List,
    ListItem,
    Loader,
    Segment
} from "semantic-ui-react";
import {CSSProperties, useContext, useEffect, useState} from "react";
import {GetMyProfile, GetMyProfileResponse, PlayerColor, SetMyProfile, SetMyProfileRequest} from "../api/api.ts";
import ErrorContext from "../ErrorContext.tsx";
import {playerColorToHtml} from "../actions/renderer/HexRenderer.tsx";
import {discordOauthSignin, googleOauthSignin} from "../UserSessionContext.tsx";

function PlayerColorDot({color, onClick}: {color: PlayerColor, onClick?: () => void}) {
    let style: CSSProperties = {
        height: '1em',
        width: '1em',
        borderRadius: '50%',
        display: 'inline-block',
        backgroundColor: playerColorToHtml(color),
    };
    if (onClick) {
        style.cursor = 'pointer';
    };
    return <div style={style} onClick={onClick}/>
}

function ProfilePage() {
    let {setError} = useContext(ErrorContext);
    let [profile, setProfile] = useState<GetMyProfileResponse|undefined>(undefined);
    let [newProfile, setNewProfile] = useState<SetMyProfileRequest|undefined>(undefined);
    let [saving, setSaving] = useState<boolean>(false);

    const reload = () => {
        GetMyProfile({}).then(res => {
            setProfile(res);
            setNewProfile({
                emailNotificationsEnabled: res.emailNotificationsEnabled,
                discordTurnAlertsEnabled: res.discordTurnAlertsEnabled,
                colorPreferences: res.colorPreferences,
                webhooks: res.webhooks,
            })
        }).catch(err => {
            setError(err);
        });
    }

    useEffect(() => {
        reload();
    }, []);

    if (!profile) {
        return <Loader active={true} />
    }

    let unselectedColors: PlayerColor[] = [];
    for (let color of [PlayerColor.BLUE, PlayerColor.GREEN, PlayerColor.YELLOW, PlayerColor.PINK, PlayerColor.GRAY, PlayerColor.ORANGE]) {
        if (!newProfile || !newProfile.colorPreferences || newProfile.colorPreferences.indexOf(color) === -1) {
            unselectedColors.push(color);
        }
    }

    return <Segment>
        <Header as='h1'>My Profile</Header>
        <Form>

            <FormField>
                <label>Nickname</label>
                <FormInput disabled value={profile.nickname} />
            </FormField>
            <FormField>
                <label>Email</label>
                <p>This is automatically determined from your linked account that you use to sign in.</p>
                <FormInput disabled value={profile.email}/>
            </FormField>
            <FormField>
                <label>Email Notifications</label>
                <p>When enabled, you will receive an email to the above address whenever it becomes your turn on any game.
                    If it is not enabled, you will only receive a summary email once a day if there are any games where it's your turn.</p>
                <Checkbox toggle checked={newProfile?.emailNotificationsEnabled} onChange={(_, val) => {
                    let newNewProfile = Object.assign({}, newProfile);
                    newNewProfile.emailNotificationsEnabled = val.checked;
                    setNewProfile(newNewProfile);
                }} />
            </FormField>
            <FormField>
                <label>Discord Notifications</label>
                <p>When enabled, a message will be sent to the shared #turn-alerts channel whenever it is your turn. Link your Discord account to get an @-mention!</p>
                <Checkbox toggle checked={newProfile?.discordTurnAlertsEnabled} onChange={(_, val) => {
                    let newNewProfile = Object.assign({}, newProfile);
                    newNewProfile.discordTurnAlertsEnabled = val.checked;
                    setNewProfile(newNewProfile);
                }} />
            </FormField>
            <Header as='h2'>Linked Accounts</Header>
            <FormField>
                <label>Google User ID</label>
                <p>If you link your Google account to your profile, you can sign in with Google..</p>
                <Button secondary disabled={!!profile.googleId} onClick={() => {
                    googleOauthSignin('/linkProfile/google');
                }}>{profile.googleId ? 'Linked' : 'Link My Google Account'}</Button>
            </FormField>
            <FormField>
                <label>Discord User ID</label>
                <p>If you link your Discord account to your profile, you can sign in with Discord and webhook notifications will mention you by your Discord handle.</p>
                <Button secondary disabled={!!profile.discordId} onClick={() => {
                    discordOauthSignin('/linkProfile/discord');
                }}>{profile.discordId ? 'Linked' : 'Link My Discord Account'}</Button>
            </FormField>
            <Header as='h2'>Color Preferences</Header>
            <FormField>
                <p>If you define color preferences here, you will be assigned the given player color when starting a
                    new game.
                    If multiple players have the same top preference, one player will be randomly chosen and the
                    next-highest preference will be used, as available.</p>
                <div>
                    My preference (left to right):<br/>
                    <Segment raised>{!newProfile?.colorPreferences ? null : newProfile.colorPreferences.map((c, idx) => {
                        return <Label as='a' onClick={() => {
                            let newNewProfile = Object.assign({}, newProfile);
                            newNewProfile.colorPreferences = newNewProfile.colorPreferences?.slice() || [];
                            newNewProfile.colorPreferences.splice(idx, 1);
                            setNewProfile(newNewProfile);
                        }}>
                            <Icon name='remove' /> <PlayerColorDot color={c} />
                        </Label>})
                    }</Segment>
                    <div>
                        {unselectedColors.map(c =>
                            <Label as='a' onClick={() => {
                                let newNewProfile = Object.assign({}, newProfile);
                                newNewProfile.colorPreferences = newNewProfile.colorPreferences?.slice() || [];
                                newNewProfile.colorPreferences.push(c);
                                setNewProfile(newNewProfile);
                            }}>
                                <Icon name='add'/> <PlayerColorDot color={c} />
                            </Label>)}
                    </div>
                </div>
            </FormField>
            <Header as='h2'>Webhooks</Header>
            <FormField>
                <p>You can add webhook URLs here to receive messages when it is your turn. At the moment, only Discord webhooks are supported.</p>
                <div>
                    <List>
                        {newProfile?.webhooks?.map((webhook, idx) => {
                            return <ListItem><Input value={webhook}
                                                    onChange={(_, data) => {
                                                        let newNewProfile = Object.assign({}, newProfile);
                                                        let webhooks = newNewProfile.webhooks?.slice() || [];
                                                        webhooks[idx] = data.value;
                                                        newNewProfile.webhooks = webhooks;
                                                        setNewProfile(newNewProfile);
                                                    }}
                                                    icon={<Icon name='delete' circular link onClick={() => {
                                                        let newNewProfile = Object.assign({}, newProfile);
                                                        let webhooks = newNewProfile.webhooks?.slice() || [];
                                                        webhooks.splice(idx, 1);
                                                        newNewProfile.webhooks = webhooks;
                                                        setNewProfile(newNewProfile);
                                                    }} />}

                            /></ListItem>
                        })}
                    </List>
                    <Button secondary icon="plus" onClick={() => {
                        let newNewProfile = Object.assign({}, newProfile);
                        let webhooks = newNewProfile.webhooks?.slice() || [];
                        webhooks.push("");
                        newNewProfile.webhooks = webhooks;
                        setNewProfile(newNewProfile);
                    }}/>
                </div>
            </FormField>
            <FormButton primary loading={saving} onClick={() => {
                if (!newProfile) {
                    return;
                }
                setSaving(true);
                SetMyProfile(newProfile).then(() => {
                    return reload();
                }).catch(err => {
                    setError(err);
                }).finally(() => {
                    setSaving(false);
                })
            }}>Save</FormButton>
        </Form>
    </Segment>
}

export default ProfilePage
