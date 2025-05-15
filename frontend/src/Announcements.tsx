import {useEffect, useState} from "react";
import {Button, Header, Modal, ModalActions, ModalContent, ModalDescription, ModalHeader} from "semantic-ui-react";

const LAST_ANNOUNCEMENT = "1747344570";

export function Announcements() {
    const [display, setDisplay] = useState<boolean>(false);

    useEffect(() => {
        let lastAnnouncementSeen = localStorage.getItem('lastAnnouncementSeen');
        if (!lastAnnouncementSeen || lastAnnouncementSeen !== LAST_ANNOUNCEMENT) {
            setDisplay(true);
        }
    }, []);

    if (!display) {
        return null;
    }

    return <Modal open={true}>
        <ModalHeader>Announcements</ModalHeader>
        <ModalContent>
            <ModalDescription>
                <Header>May 15, 2025: Age of Steam Expansion Volumes V &amp; VI Kickstarter</Header>
                <img style={{maxWidth: "100%"}} src="https://i.kickstarter.com/assets/049/283/980/669e04f0c53a9f5787c0a59c3daa5a91_original.png?anim=false&fit=cover&gravity=auto&height=576&origin=ugc&q=92&v=1747277579&width=1024&sig=Fp09ai4SX%2BQpilj3iuronHf8cNBJ42gA1VqlFoa2f7U%3D" />
                <p>Eagle-Gryphon Games has launched a kickstarter for Age of Steam expansion volumes V and VI! The kickstarter includes 16 new deluxe edition maps, new display boards, and more! Learn more at the <a target="_blank" href="https://www.kickstarter.com/projects/eaglegryphon/age-of-steam-deluxe-expansion-2025">kickstarter page</a>. Please consider supporting the publisher and map designers with a pledge!</p>
            </ModalDescription>
        </ModalContent>
        <ModalActions>
            <Button primary onClick={() => {
                localStorage.setItem('lastAnnouncementSeen', LAST_ANNOUNCEMENT);
                setDisplay(false);
            }}>Dismiss</Button>
        </ModalActions>
    </Modal>
}
