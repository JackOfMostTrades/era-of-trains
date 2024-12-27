import {
    Form,
    FormButton,
    FormField,
    FormGroup,
    FormInput,
    Header,
    Icon,
    Label,
    Loader,
    Segment
} from "semantic-ui-react";
import {CSSProperties, useContext, useEffect, useState} from "react";
import {GetMyProfile, GetMyProfileResponse, PlayerColor, SetMyProfile, SetMyProfileRequest} from "../api/api.ts";
import ErrorContext from "../ErrorContext.tsx";
import {playerColorToHtml} from "../actions/renderer/HexRenderer.tsx";

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
                colorPreferences: res.colorPreferences
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
            <FormGroup widths='equal'>
                <FormField>
                    <label>Nickname</label>
                    <FormInput disabled value={profile.nickname} />
                </FormField>
                <FormField>
                    <label>Email</label>
                    <FormInput disabled value={profile.email} />
                </FormField>
            </FormGroup>
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
