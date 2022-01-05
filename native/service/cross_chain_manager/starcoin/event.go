/** The  poly network  is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
* GNU Lesser General Public License for more details.
* You should have received a copy of the GNU Lesser General Public License
* along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */

package starcoin

import (
	"fmt"

	"github.com/novifinancial/serde-reflection/serde-generate/runtime/golang/bcs"
	"github.com/novifinancial/serde-reflection/serde-generate/runtime/golang/serde"
)

type CrossChainEvent struct {
	Sender               []byte
	TxId                 []byte
	ProxyOrAssetContract []byte
	ToChainId            uint64
	ToContract           []byte
	RawData              []byte
}

func (obj *CrossChainEvent) Serialize(serializer serde.Serializer) error {
	if err := serializer.IncreaseContainerDepth(); err != nil {
		return err
	}
	if err := serializer.SerializeBytes(obj.Sender); err != nil {
		return err
	}
	if err := serializer.SerializeBytes(obj.TxId); err != nil {
		return err
	}
	if err := serializer.SerializeBytes(obj.ProxyOrAssetContract); err != nil {
		return err
	}
	if err := serializer.SerializeU64(obj.ToChainId); err != nil {
		return err
	}
	if err := serializer.SerializeBytes(obj.ToContract); err != nil {
		return err
	}
	if err := serializer.SerializeBytes(obj.RawData); err != nil {
		return err
	}
	serializer.DecreaseContainerDepth()
	return nil
}

func (obj *CrossChainEvent) BcsSerialize() ([]byte, error) {
	if obj == nil {
		return nil, fmt.Errorf("Cannot serialize null object")
	}
	serializer := bcs.NewSerializer()
	if err := obj.Serialize(serializer); err != nil {
		return nil, err
	}
	return serializer.GetBytes(), nil
}

func DeserializeCrossChainEvent(deserializer serde.Deserializer) (CrossChainEvent, error) {
	var obj CrossChainEvent
	if err := deserializer.IncreaseContainerDepth(); err != nil {
		return obj, err
	}
	if val, err := deserializer.DeserializeBytes(); err == nil {
		obj.Sender = val
	} else {
		return obj, err
	}
	if val, err := deserializer.DeserializeBytes(); err == nil {
		obj.TxId = val
	} else {
		return obj, err
	}
	if val, err := deserializer.DeserializeBytes(); err == nil {
		obj.ProxyOrAssetContract = val
	} else {
		return obj, err
	}
	if val, err := deserializer.DeserializeU64(); err == nil {
		obj.ToChainId = val
	} else {
		return obj, err
	}
	if val, err := deserializer.DeserializeBytes(); err == nil {
		obj.ToContract = val
	} else {
		return obj, err
	}
	if val, err := deserializer.DeserializeBytes(); err == nil {
		obj.RawData = val
	} else {
		return obj, err
	}
	deserializer.DecreaseContainerDepth()
	return obj, nil
}

func BcsDeserializeCrossChainEvent(input []byte) (CrossChainEvent, error) {
	if input == nil {
		var obj CrossChainEvent
		return obj, fmt.Errorf("Cannot deserialize null array")
	}
	deserializer := bcs.NewDeserializer(input)
	obj, err := DeserializeCrossChainEvent(deserializer)
	if err == nil && deserializer.GetBufferOffset() < uint64(len(input)) {
		return obj, fmt.Errorf("Some input bytes were not read")
	}
	return obj, err
}
